cmake_minimum_required(VERSION 3.20)

project(Taco VERSION 0.1)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED true)

# Set the output directory for binaries (executables) directly in the root directory
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_SOURCE_DIR}/bin)

# Create a custom target to ensure the directories are created before build
file(MAKE_DIRECTORY ${CMAKE_SOURCE_DIR}/bin)

find_package(Curses REQUIRED)
include_directories(${CURSES_INCLUDE_DIR})

include_directories(include)

set(SOURCE_FILES
    src/main.cpp
    src/Session.cpp
    src/Menu.cpp
)

add_executable(taco ${SOURCE_FILES})

target_link_libraries(taco ${CURSES_LIBRARIES})