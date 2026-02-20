cmake_minimum_required(VERSION 3.25)
project(pganalytics-collector VERSION 3.0.0 LANGUAGES CXX)

# Set C++ standard
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fPIC -Wall -Wextra -pedantic")

# Option for building tests
option(BUILD_TESTS "Build unit tests" ON)

# Find required packages
find_package(OpenSSL 3.0 REQUIRED)
find_package(CURL REQUIRED)
find_package(PostgreSQL)
find_package(ZLIB REQUIRED)
find_package(nlohmann_json 3.2.0 QUIET)

# If nlohmann_json not found, include it as header-only
if(NOT nlohmann_json_FOUND)
    include_directories(${CMAKE_CURRENT_SOURCE_DIR}/include/nlohmann)
endif()

# Check for libpq availability
if(PostgreSQL_FOUND)
    message(STATUS "PostgreSQL found: ${PostgreSQL_VERSION_STRING}")
    message(STATUS "PostgreSQL include: ${PostgreSQL_INCLUDE_DIRS}")
    message(STATUS "PostgreSQL libraries: ${PostgreSQL_LIBRARIES}")
    add_compile_definitions(HAVE_LIBPQ)
else()
    message(STATUS "PostgreSQL not found - pg_stats collector will use default values")
endif()

# Include directories
include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}/include
    ${OPENSSL_INCLUDE_DIR}
    ${CURL_INCLUDE_DIRS}
    ${ZLIB_INCLUDE_DIRS}
)

# Add PostgreSQL include dirs if found
if(PostgreSQL_FOUND)
    include_directories(${PostgreSQL_INCLUDE_DIRS})
endif()

# Source files
set(COLLECTOR_SOURCES
    src/main.cpp
    src/collector.cpp
    src/postgres_plugin.cpp
    src/sysstat_plugin.cpp
    src/log_plugin.cpp
    src/sender.cpp
    src/auth.cpp
    src/config_manager.cpp
    src/metrics_serializer.cpp
    src/metrics_buffer.cpp
)

# Header files
set(COLLECTOR_HEADERS
    include/collector.h
    include/postgres_plugin.h
    include/sysstat_plugin.h
    include/log_plugin.h
    include/sender.h
    include/auth.h
    include/config_manager.h
    include/metrics_serializer.h
    include/metrics_buffer.h
)

# Create executable
add_executable(pganalytics ${COLLECTOR_SOURCES} ${COLLECTOR_HEADERS})

# Link libraries
target_link_libraries(pganalytics
    PRIVATE
    OpenSSL::SSL OpenSSL::Crypto
    ${CURL_LIBRARIES}
    ${ZLIB_LIBRARIES}
    pthread
)

# Add PostgreSQL libraries if found
if(PostgreSQL_FOUND)
    target_link_libraries(pganalytics PRIVATE ${PostgreSQL_LIBRARIES})
endif()

# Compiler flags
if(CMAKE_CXX_COMPILER_ID MATCHES "GNU|Clang")
    target_compile_options(pganalytics PRIVATE -Wall -Wextra -Wpedantic -Werror=format-security)
endif()

# Set output directory
set_target_properties(pganalytics PROPERTIES
    RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/src"
)

# Tests
if(BUILD_TESTS)
    enable_testing()
    add_subdirectory(tests)
endif()

# Installation
install(TARGETS pganalytics DESTINATION bin)
install(FILES config.toml.sample DESTINATION etc/pganalytics)

# Print configuration summary
message(STATUS "pgAnalytics Collector v${PROJECT_VERSION}")
message(STATUS "Build type: ${CMAKE_BUILD_TYPE}")
message(STATUS "C++ compiler: ${CMAKE_CXX_COMPILER}")
message(STATUS "OpenSSL: ${OPENSSL_VERSION}")
message(STATUS "CURL: ${CURL_VERSION_STRING}")
message(STATUS "PostgreSQL: ${PostgreSQL_VERSION_STRING}")
