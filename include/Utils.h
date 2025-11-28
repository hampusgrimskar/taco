#ifndef UTILS
#define UTILS

#include <stdlib.h>
#include <iomanip>
#include <sstream>
#include <unordered_map>
#include <fstream>
#include <filesystem>
#include "cxxopts.hpp"
#include "Constants.h"
#include <chrono>

inline void exitWithError(const std::string &message)
{
    std::cout << "Error: " + message << std::endl;
    exit(EXIT_FAILURE);
}

inline void exitWithMessage(const std::string &message, const int status_code)
{
    std::cout << message << std::endl;
    exit(status_code);
}

inline std::string generateId()
{
    std::stringstream output_stream;

    // Add timestamp to ID to prevent creating similar IDs
    auto now = std::chrono::system_clock::now();
    auto millis = std::chrono::duration_cast<std::chrono::milliseconds>(now.time_since_epoch()).count();
    output_stream << millis;
    return output_stream.str();
}

inline void initializeConfigurationFiles()
{
    // Create .taco directory and config file if they don't exist already
    if (!std::filesystem::exists(TACO_CONFIG_DIR))
    {
        std::filesystem::create_directory(TACO_CONFIG_DIR);
    }
    if (!std::ifstream(TACO_CONFIG_FILE))
    {
        std::ofstream file(TACO_CONFIG_FILE);
    }
}

inline std::unordered_map<std::string, std::string> getReposFromConfigFile()
{
    std::ifstream file(TACO_CONFIG_FILE);
    std::string line;
    std::unordered_map<std::string, std::string> repos;
    while (getline(file, line))
    {
        std::string path;
        std::string alias;

        int delimiter_index = line.find('#');
        if (delimiter_index != -1)
        {
            path = line.substr(0, delimiter_index);
            alias = line.substr(delimiter_index + 1, line.length());
        }
        else
        {
            path = line;
            alias = NO_ALIAS;
        }

        repos.emplace(path, alias);
    }
    file.close();
    return repos;
}

inline void updateConfigurationFile(std::unordered_map<std::string, std::string> updated_config)
{
    std::ofstream config_file(TACO_CONFIG_FILE, std::ios::trunc);

    for (const auto& [path, alias] : updated_config)
    {
        config_file << path + "#" + alias << std::endl;
    }

    config_file.close();
}

inline void updateAliasInConfig(std::string repository_path, std::string new_alias)
{
    std::unordered_map<std::string, std::string> config = getReposFromConfigFile();
    config[repository_path] = new_alias;
    updateConfigurationFile(config);
}

inline bool isRepoInitialized(const std::string &path)
{
    std::ifstream file(TACO_CONFIG_FILE);
    std::string line;
    while (getline(file, line))
    {
        int delimiter_index = line.find('#');
        if (delimiter_index != -1)
            line = line.substr(0, delimiter_index);

        if (line == path)
        {
            file.close();
            return true;
        }
    }
    file.close();
    return false;
}

inline void initializeRepository(const cxxopts::ParseResult &result)
{
    std::string pwd = std::getenv("PWD");
    std::ofstream file(TACO_CONFIG_FILE, std::ios::app);

    if (isRepoInitialized(pwd))
        exitWithError("directory " + pwd + " is already initialized!");

    std::string message;

    // Add alias if alias option was set
    if (result.contains("alias"))
    {
        std::string alias = result["alias"].as<std::string>();
        file << pwd + "#" + alias << std::endl;
        message = "Succesfully initialized repository " + pwd + " with alias " + alias;
    }

    else
    {
        file << pwd << std::endl;
        message = "Succesfully initialized repository " + pwd;
    }

    file.close();

    exitWithMessage(message, EXIT_SUCCESS);
}

#endif
