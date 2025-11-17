#include "Session.h"
#include <fstream>
#include <filesystem>
#include <cstdlib>
#include <string>
#include <cstring>
#include <csignal>
#include "cxxopts.hpp"
// #include "Menu.h"
#include "FtxMenu.h"

// Global menu object
// Menu *MAIN_MENU = new Menu();
FtxMenu *MAIN_MENU = new FtxMenu();

/*
Probably should move some stuff from
this file into Utils.h
*/

void handleArguments(int argc, char **argv)
{
    cxxopts::Options options("taco", "Repository navigator");

    options.add_options()
    ("h,help", "Show help")
    ("i,init", "Initialize a directory for taco use")
    ("a,alias", "Alias to use instead of full path to repository. Can only be used together with init option",cxxopts::value<std::string>())
    ("rm,remove", "Remove this repository from taco (Note: This will NOT change or remove any files in the repo)");

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

    if (result["remove"].as<bool>())
    {
        // MAIN_MENU->removeRepository(result);
    }
}

void signalHandler(int signal)
{
    if (signal == SIGINT)
    {
        // Clean up sessions on SIGINT signal
        MAIN_MENU->~FtxMenu();
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
        MAIN_MENU->show();
    }

    MAIN_MENU->~FtxMenu();

    return 0;
}
