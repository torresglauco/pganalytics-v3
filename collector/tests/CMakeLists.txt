# Tests CMakeLists.txt

# Find Google Test
find_package(GTest REQUIRED)

# Unit test sources
set(UNIT_TEST_SOURCES
    unit/metrics_serializer_test.cpp
    unit/auth_test.cpp
    unit/metrics_buffer_test.cpp
    unit/config_manager_test.cpp
    unit/sender_test.cpp
    postgres_plugin_test.cpp
)

# Integration test sources
set(INTEGRATION_TEST_SOURCES
    integration/mock_backend_server.cpp
    integration/sender_integration_test.cpp
    integration/collector_flow_test.cpp
    integration/auth_integration_test.cpp
    integration/config_integration_test.cpp
    integration/error_handling_test.cpp
)

# E2E test sources
set(E2E_TEST_SOURCES
    e2e/e2e_harness.cpp
    e2e/http_client.cpp
    e2e/database_helper.cpp
    e2e/grafana_helper.cpp
    e2e/1_collector_registration_test.cpp
    e2e/2_metrics_ingestion_test.cpp
    e2e/3_configuration_test.cpp
    e2e/4_dashboard_visibility_test.cpp
    e2e/5_performance_test.cpp
    e2e/6_failure_recovery_test.cpp
)

# Combine all test sources
set(TEST_SOURCES
    ${UNIT_TEST_SOURCES}
    ${INTEGRATION_TEST_SOURCES}
    ${E2E_TEST_SOURCES}
)

# Exclude main.cpp from tests
set(COLLECTOR_SOURCES_NO_MAIN
    ${CMAKE_CURRENT_SOURCE_DIR}/../src/collector.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/../src/postgres_plugin.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/../src/sysstat_plugin.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/../src/log_plugin.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/../src/sender.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/../src/auth.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/../src/config_manager.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/../src/metrics_serializer.cpp
    ${CMAKE_CURRENT_SOURCE_DIR}/../src/metrics_buffer.cpp
)

# Create test executable
add_executable(pganalytics-tests
    ${TEST_SOURCES}
    ${COLLECTOR_SOURCES_NO_MAIN}
)

# Include directories for tests
target_include_directories(pganalytics-tests PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/../include
    ${CMAKE_CURRENT_SOURCE_DIR}/integration
    ${CMAKE_CURRENT_SOURCE_DIR}/e2e
    ${GTEST_INCLUDE_DIRS}
    ${nlohmann_json_INCLUDE_DIR}
)

# Link libraries
target_link_libraries(pganalytics-tests
    PRIVATE
    GTest::gtest
    GTest::gtest_main
    OpenSSL::SSL OpenSSL::Crypto
    ${CURL_LIBRARIES}
    ${ZLIB_LIBRARIES}
    pthread
)

# Add PostgreSQL libraries if found
if(PostgreSQL_FOUND)
    target_link_libraries(pganalytics-tests PRIVATE ${PostgreSQL_LIBRARIES})
endif()

# Compiler flags
if(CMAKE_CXX_COMPILER_ID MATCHES "GNU|Clang")
    target_compile_options(pganalytics-tests PRIVATE -Wall -Wextra -Wpedantic)
endif()

# Register tests
gtest_discover_tests(pganalytics-tests)

# Set output directory
set_target_properties(pganalytics-tests PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/tests"
)

# Optional: Code coverage target
if(CMAKE_CXX_COMPILER_ID STREQUAL "GNU")
    find_program(GCOV gcov)
    find_program(LCOV lcov)

    if(GCOV AND LCOV)
        message(STATUS "Code coverage tools found - 'make coverage' target available")

        add_custom_target(coverage
            COMMAND ${CMAKE_CXXFLAGS} --cov
            COMMAND ./pganalytics-tests
            WORKING_DIRECTORY ${CMAKE_BINARY_DIR}/tests
            COMMENT "Running tests with coverage"
        )
    endif()
endif()

message(STATUS "Tests configured - run with: ctest or ./pganalytics-tests")
