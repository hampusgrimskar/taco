#ifndef MENU
#define MENU

#include <ncurses.h>
#include <stdio.h>
#include <vector>
#include <string>
#include "Session.h"

class Menu
{
private:
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

        std::string getSessionNameWithStatus()
        {
            return (this->session_name + (is_active ? " (Active)" : ""));
        }
    };
    

    // std::vector<std::tuple<std::string, std::string, Session *>> repositorySessions;
    std::vector<Menu::RepositorySession> repositorySessions;

    void printTitle();

    void printMenu(WINDOW *menu_win, int highlight);

    void handleSelection(int selection);

public:
    Menu();

    ~Menu();

    void openMenu();
};

#endif
