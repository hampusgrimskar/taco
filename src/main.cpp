#include "Session.h"
#include <fstream>
#include <filesystem>
#include <cstdlib>
#include <string>
#include <cstring>
#include "cxxopts.hpp"

const std::string FORWARD_SLASH = "/";
const std::string DOT = ".";
const std::string NO_ALIAS = "";

std::string HOME_DIR = std::getenv("HOME");
std::string TACO_CONFIG_DIR = HOME_DIR + FORWARD_SLASH + DOT + "taco/";
std::string TACO_CONFIG_FILE = TACO_CONFIG_DIR + "repositories";

void initializeConfigurationFiles()
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

std::unordered_map<std::string, std::string> getReposFromConfigFile()
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

bool isRepoInitialized(const std::string &path)
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

void exitWithError(const std::string &message)
{
    std::cout << "Error: " + message << std::endl;
    exit(EXIT_FAILURE);
}

void handleInit(const cxxopts::ParseResult &result)
{
    std::string pwd = std::getenv("PWD");
    std::ofstream file(TACO_CONFIG_FILE, std::ios::app);

    if (isRepoInitialized(pwd))
        exitWithError("directory " + pwd + " is already initialized!");

    // Add alias if alias option was set
    if (result.contains("alias"))
    {
        file << pwd + "#" + result["alias"].as<std::string>() << std::endl;
    }

    else
    {
        file << pwd << std::endl;
    }

    file.close();
}

std::pair<std::string, std::string> chooseRepo()
{
    // TEMPORARY CODE FOR TESTING
    std::cout << "\nTaco main session running...\ntype q to quit." << std::endl;
    int menu_counter = 0;
    std::vector<std::pair<std::string, std::string>> options;

    for (auto repo : getReposFromConfigFile())
    {
        std::string name = repo.second != NO_ALIAS ? repo.second : repo.first;
        std::cout << menu_counter++ << ": " + name << std::endl;
        options.push_back(std::pair(repo.first, repo.second));
    }
    int choice = std::stoi(std::string(1, getchar()));
    return options[choice];
}

void handleArguments(int argc, char **argv)
{
    cxxopts::Options options("taco", "Repository navigator");

    options.add_options()
    ("h,help", "Show help")
    ("i,init", "Initialize a directory for taco use")
    ("a,alias", "Alias to use instead of full path to repository. Can only be used together with init option",
    cxxopts::value<std::string>());

    cxxopts::ParseResult result = options.parse(argc, argv);

    if (result["help"].as<bool>())
    {
        std::cout << options.help() << std::endl;
        exit(0);
    }

    if (result["init"].as<bool>())
    {
        handleInit(result);
    }
}

int main(int argc, char *argv[])
{
    initializeConfigurationFiles();

    handleArguments(argc, argv);

    Session* session;
    bool is_session_created = false;

    std::cout << "\nTaco main session running...\ntype q to quit." << std::endl;
    while (true)
    {
        try
        {
            std::pair<std::string, std::string> repo = chooseRepo();

            std::string session_name = repo.second != NO_ALIAS ? repo.second : repo.first;
            
            if (is_session_created) session->attach();
            else {
                session = new Session(session_name.c_str(), repo.first.c_str());
                is_session_created = true;
            } 

            if (getchar() == 'q')
                break;
        }
        catch(const std::exception& e)
        {
            delete session;
        }
    }
    delete session;
    return 0;
}