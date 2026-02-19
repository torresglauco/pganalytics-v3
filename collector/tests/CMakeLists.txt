# Tests CMakeLists.txt

# Find Google Test
find_package(GTest REQUIRED)

# Test sources
set(TEST_SOURCES
    unit/metrics_serializer_test.cpp
    unit/auth_test.cpp
    unit/metrics_buffer_test.cpp
    unit/config_manager_test.cpp
    unit/sender_test.cpp
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
    ${GTEST_INCLUDE_DIRS}
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
