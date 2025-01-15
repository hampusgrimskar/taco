#include "Session.h"
#include <fstream>
#include <filesystem>
#include <cstdlib>
#include <string>
#include <cstring>
#include <csignal>
#include "cxxopts.hpp"
#include "Menu.h"

// Global menu object
Menu *MAIN_MENU = new Menu();

/*
Probably should move some stuff from
this file into Utils.h
*/

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

    options.add_options()("h,help", "Show help")("i,init", "Initialize a directory for taco use")("a,alias", "Alias to use instead of full path to repository. Can only be used together with init option",
                                                                                                  cxxopts::value<std::string>());

    cxxopts::ParseResult result = options.parse(argc, argv);

    if (result["help"].as<bool>())
    {
        std::cout << options.help() << std::endl;
        exit(0);
    }

    if (result["init"].as<bool>())
    {
        initializeRepository(result);
    }
}

void signalHandler(int signal)
{
    if (signal == SIGINT)
    {
        // Clean up sessions on SIGINT signal
        MAIN_MENU->~Menu();
        exit(0);
    }
}

int main(int argc, char *argv[])
{
    initializeConfigurationFiles();

    handleArguments(argc, argv);

    // Cleanup Session instances when SIGINT
    // signal is received
    signal(SIGINT, signalHandler);

    while (1)
    {
        MAIN_MENU->openMenu();
    }

    MAIN_MENU->~Menu();

    return 0;
}
