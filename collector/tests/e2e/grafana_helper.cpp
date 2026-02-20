#include "grafana_helper.h"
#include <curl/curl.h>
#include <iostream>

// CURL callback for response body
static size_t writeCallback(void* contents, size_t size, size_t nmemb, std::string* userp) {
    userp->append((char*)contents, size * nmemb);
    return size * nmemb;
}

E2EGrafanaHelper::E2EGrafanaHelper(
    const std::string& grafana_url,
    const std::string& api_key
)
    : m_grafana_url(grafana_url),
      m_api_key(api_key),
      m_verbose(false) {
}

E2EGrafanaHelper::~E2EGrafanaHelper() {
}

bool E2EGrafanaHelper::isHealthy() {
    return testConnection();
}

bool E2EGrafanaHelper::testConnection() {
    std::string response_body;
    int response_code = 0;

    bool success = performRequest("GET", "/api/health", "", response_body, response_code);

    if (m_verbose) {
        std::cout << "[E2E Grafana] Health check: " << response_code << std::endl;
    }

    return success && response_code == 200;
}

bool E2EGrafanaHelper::isDatasourceHealthy(const std::string& datasource_name) {
    // List all datasources
    std::string response_body;
    int response_code = 0;

    if (!performRequest("GET", "/api/datasources", "", response_body, response_code)) {
        return false;
    }

    // Check if datasource exists and is healthy
    return response_body.find(datasource_name) != std::string::npos &&
           response_body.find("\"isDefault\"") != std::string::npos;
}

std::vector<std::string> E2EGrafanaHelper::listDatasources() {
    std::vector<std::string> datasources;
    std::string response_body;
    int response_code = 0;

    if (performRequest("GET", "/api/datasources", "", response_body, response_code)) {
        parseJsonArray(response_body, datasources);
    }

    return datasources;
}

std::string E2EGrafanaHelper::getDatasourceStatus(const std::string& name) {
    // Query datasource health
    std::string response_body;
    int response_code = 0;

    std::string endpoint = "/api/datasources/name/" + name;
    if (performRequest("GET", endpoint, "", response_body, response_code)) {
        return getJsonField(response_body, "\"type\"");
    }

    return "";
}

bool E2EGrafanaHelper::dashboardExists(const std::string& dashboard_uid) {
    std::string response_body;
    int response_code = 0;

    std::string endpoint = "/api/dashboards/uid/" + dashboard_uid;
    return performRequest("GET", endpoint, "", response_body, response_code) &&
           response_code == 200;
}

std::string E2EGrafanaHelper::getDashboard(const std::string& dashboard_uid) {
    std::string response_body;
    int response_code = 0;

    std::string endpoint = "/api/dashboards/uid/" + dashboard_uid;
    if (performRequest("GET", endpoint, "", response_body, response_code)) {
        return response_body;
    }

    return "";
}

std::vector<std::string> E2EGrafanaHelper::listDashboards() {
    std::vector<std::string> dashboards;
    std::string response_body;
    int response_code = 0;

    if (performRequest("GET", "/api/search", "", response_body, response_code)) {
        parseJsonArray(response_body, dashboards);
    }

    return dashboards;
}

bool E2EGrafanaHelper::dashboardLoads(const std::string& dashboard_uid) {
    std::string dashboard = getDashboard(dashboard_uid);
    return !dashboard.empty() && dashboard.find("\"dashboard\"") != std::string::npos;
}

bool E2EGrafanaHelper::panelDataAvailable(const std::string& dashboard_uid, int panel_id) {
    std::string dashboard = getDashboard(dashboard_uid);

    if (dashboard.empty()) {
        return false;
    }

    // Check if panel exists in dashboard
    std::string panel_search = "\"id\":" + std::to_string(panel_id);
    return dashboard.find(panel_search) != std::string::npos;
}

std::string E2EGrafanaHelper::getPanelData(const std::string& dashboard_uid, int panel_id) {
    // Get dashboard and find panel
    std::string dashboard = getDashboard(dashboard_uid);

    if (dashboard.empty()) {
        m_last_error = "Dashboard not found";
        return "";
    }

    // Extract panel data (simplified - real implementation would parse JSON)
    return dashboard;
}

std::vector<std::string> E2EGrafanaHelper::listAlerts() {
    std::vector<std::string> alerts;
    std::string response_body;
    int response_code = 0;

    if (performRequest("GET", "/api/alerts", "", response_body, response_code)) {
        parseJsonArray(response_body, alerts);
    }

    return alerts;
}

std::string E2EGrafanaHelper::getAlertStatus(const std::string& alert_uid) {
    std::string response_body;
    int response_code = 0;

    std::string endpoint = "/api/alerts/uid/" + alert_uid;
    if (performRequest("GET", endpoint, "", response_body, response_code)) {
        return getJsonField(response_body, "\"state\"");
    }

    return "";
}

bool E2EGrafanaHelper::alertExists(const std::string& alert_name) {
    std::vector<std::string> alerts = listAlerts();

    for (const auto& alert : alerts) {
        if (alert.find(alert_name) != std::string::npos) {
            return true;
        }
    }

    return false;
}

bool E2EGrafanaHelper::isAlertFiring(const std::string& alert_name) {
    std::string response_body;
    int response_code = 0;

    if (performRequest("GET", "/api/alerts", "", response_body, response_code)) {
        // Look for alert with "firing" state
        size_t pos = response_body.find(alert_name);
        if (pos != std::string::npos) {
            // Check if "state" is "firing" near this position
            size_t state_pos = response_body.find("\"state\"", pos);
            if (state_pos != std::string::npos) {
                return response_body.find("\"firing\"", state_pos) < state_pos + 50;
            }
        }
    }

    return false;
}

std::string E2EGrafanaHelper::executeQuery(
    const std::string& datasource_name,
    const std::string& query,
    int time_range_seconds
) {
    std::string response_body;
    int response_code = 0;

    // Build query request
    std::string endpoint = "/api/datasources/proxy/1/query";
    std::string body = "{\"queries\":[{\"datasource\":{\"name\":\"" + datasource_name +
                       "\"},\"query\":\"" + query + "\"}]}";

    if (performRequest("POST", endpoint, body, response_body, response_code)) {
        return response_body;
    }

    return "";
}

void E2EGrafanaHelper::setVerbose(bool verbose) {
    m_verbose = verbose;
}

std::string E2EGrafanaHelper::getLastError() const {
    return m_last_error;
}

bool E2EGrafanaHelper::performRequest(
    const std::string& method,
    const std::string& endpoint,
    const std::string& body,
    std::string& response_body,
    int& response_code
) {
    CURL* curl = curl_easy_init();
    if (!curl) {
        m_last_error = "Failed to initialize CURL";
        return false;
    }

    std::string full_url = m_grafana_url + endpoint;

    curl_easy_setopt(curl, CURLOPT_URL, full_url.c_str());
    curl_easy_setopt(curl, CURLOPT_HTTPGET, 1L);

    // Set method
    if (method == "POST") {
        curl_easy_setopt(curl, CURLOPT_POST, 1L);
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, body.c_str());
    }

    // Set response callback
    response_body.clear();
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, writeCallback);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &response_body);

    // Set timeout
    curl_easy_setopt(curl, CURLOPT_TIMEOUT, 10L);

    // Add headers
    struct curl_slist* headers = nullptr;
    headers = curl_slist_append(headers, "Content-Type: application/json");

    if (!m_api_key.empty()) {
        std::string auth = "Authorization: Bearer " + m_api_key;
        headers = curl_slist_append(headers, auth.c_str());
    }

    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);

    // Perform request
    CURLcode res = curl_easy_perform(curl);

    // Get response code
    long http_code = 0;
    curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &http_code);
    response_code = http_code;

    // Cleanup
    if (headers) {
        curl_slist_free_all(headers);
    }
    curl_easy_cleanup(curl);

    if (res != CURLE_OK) {
        m_last_error = curl_easy_strerror(res);
        return false;
    }

    return http_code >= 200 && http_code < 300;
}

bool E2EGrafanaHelper::parseJsonArray(const std::string& json, std::vector<std::string>& items) {
    // Simple JSON parsing (find all items in array)
    // Real implementation would use proper JSON parser
    size_t pos = 0;

    while ((pos = json.find("\"name\":\"", pos)) != std::string::npos) {
        pos += 8;
        size_t end = json.find("\"", pos);
        if (end != std::string::npos) {
            items.push_back(json.substr(pos, end - pos));
        }
    }

    return !items.empty();
}

std::string E2EGrafanaHelper::getJsonField(const std::string& json, const std::string& field) {
    size_t pos = json.find(field);
    if (pos == std::string::npos) {
        return "";
    }

    // Find colon and extract value
    pos = json.find(":", pos);
    if (pos == std::string::npos) {
        return "";
    }

    pos++;  // Skip colon
    while (pos < json.length() && (json[pos] == ' ' || json[pos] == '"')) {
        pos++;
    }

    size_t end = pos;
    while (end < json.length() && json[end] != ',' && json[end] != '}' && json[end] != '"') {
        end++;
    }

    return json.substr(pos, end - pos);
}

