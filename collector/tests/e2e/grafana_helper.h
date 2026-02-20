#pragma once

#include <string>
#include <vector>
#include <map>

/**
 * E2E Grafana Helper
 *
 * Provides utilities for testing Grafana dashboards and alerts.
 * Handles:
 * - Datasource verification
 * - Dashboard loading and queries
 * - Alert rule checking
 * - Panel data retrieval
 */
class E2EGrafanaHelper {
public:
    /**
     * Constructor
     * @param grafana_url Grafana API URL (e.g., http://localhost:3000)
     * @param api_key Grafana API key (optional, for authentication)
     */
    E2EGrafanaHelper(
        const std::string& grafana_url,
        const std::string& api_key = ""
    );

    ~E2EGrafanaHelper();

    // Health and connectivity
    bool isHealthy();
    bool testConnection();

    // Datasource operations
    bool isDatasourceHealthy(const std::string& datasource_name);
    std::vector<std::string> listDatasources();
    std::string getDatasourceStatus(const std::string& name);

    // Dashboard operations
    bool dashboardExists(const std::string& dashboard_uid);
    std::string getDashboard(const std::string& dashboard_uid);
    std::vector<std::string> listDashboards();
    bool dashboardLoads(const std::string& dashboard_uid);

    // Panel data retrieval
    bool panelDataAvailable(const std::string& dashboard_uid, int panel_id);
    std::string getPanelData(const std::string& dashboard_uid, int panel_id);

    // Alert operations
    std::vector<std::string> listAlerts();
    std::string getAlertStatus(const std::string& alert_uid);
    bool alertExists(const std::string& alert_name);
    bool isAlertFiring(const std::string& alert_name);

    // Query execution
    std::string executeQuery(
        const std::string& datasource_name,
        const std::string& query,
        int time_range_seconds = 3600
    );

    // Logging and debugging
    void setVerbose(bool verbose);
    std::string getLastError() const;

private:
    std::string m_grafana_url;
    std::string m_api_key;
    bool m_verbose;
    std::string m_last_error;

    // Internal HTTP methods
    bool performRequest(
        const std::string& method,
        const std::string& endpoint,
        const std::string& body,
        std::string& response_body,
        int& response_code
    );

    // Response parsing helpers
    bool parseJsonArray(const std::string& json, std::vector<std::string>& items);
    std::string getJsonField(const std::string& json, const std::string& field);
};

