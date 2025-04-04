#ifndef MENU
#define MENU

#include <ncurses.h>
#include <stdio.h>
#include <vector>
#include <string>
#include <algorithm>
#include "Session.h"

class Menu
{
private:
    int highlight = 0;

    struct RepositorySession
        {
            std::string session_name;
            std::string path;
            Session* session;
            bool is_active;

            RepositorySession(std::string session_name, std::string path)
            {
                this->session_name = session_name;
                this->path = path;
                this->session = nullptr;
                this->is_active = false;
            }

            bool operator==(const RepositorySession& other)
            {
                return (this->session_name == other.session_name)
                    && (this->path == other.path);
            }
        };

    int longest_title;

    void updateLongestTitle(const std::string& session_name)
    {
        if (session_name.length() > this->longest_title)
        {
            longest_title = session_name.length();
        }
    }

    std::string getSessionNameWithStatus(Menu::RepositorySession& repositorySession)
    {
        int number_of_spaces = longest_title - repositorySession.session_name.length() + 1;
        std::string spaces = std::string(number_of_spaces, ' ');
        return (repositorySession.session_name + (repositorySession.is_active ? spaces + " (Active)" : spaces + "         "));
    }

    std::vector<Menu::RepositorySession> repositorySessions;

    void printTitle();

    void sortRepositorySessions();

    void printMenu(WINDOW *menu_win);

    void handleSelection(int selection);

    void updateRepositorySessions();
    
public:
    Menu();
    
    ~Menu();
    
    void openMenu();

    void removeRepositorySession(std::string session_name);

    void removeRepository(const cxxopts::ParseResult &result);
};

#endif
