cmake_minimum_required(VERSION 3.20)

project(Taco VERSION 0.1)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED true)

include_directories(include)

set(SOURCE_FILES
    src/main.cpp
    src/Session.cpp
)

add_executable(taco ${SOURCE_FILES})