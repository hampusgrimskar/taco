#include "Session.h"
#include <fstream>
#include <filesystem>
#include <cstdlib>
#include <string>
#include <cstring>

std::string FORWARD_SLASH = "/";
std::string DOT = ".";
std::string USER = std::getenv("USER");
std::string TACO_CONFIG_DIR = FORWARD_SLASH + "home" + FORWARD_SLASH + USER + FORWARD_SLASH + DOT + "taco/";
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

void handleArguments(int argc, char** argv)
{ 
    if (argc == 2 && std::strcmp(argv[1], "init") == 0)
    {
        std::string pwd = std::getenv("PWD");
        std::ofstream file(TACO_CONFIG_FILE, std::ios::app);
        file << pwd << std::endl;
        file.close();
    }
}

int main(int argc, char* argv[]) {

    initializeConfigurationFiles();

    handleArguments(argc, argv);

    Session session("test", "~/repos");
    // session.detach();

    std::cout << "\nTaco main session running...\ntype q to quit." << std::endl;
    while(true)
    {
        if (getchar() == 'q') break;
    }
    return 0;
}