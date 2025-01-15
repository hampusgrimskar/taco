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
    std::vector<std::tuple<std::string, std::string, Session*>> repositorySessionMap;

    void printMenu(WINDOW *menu_win, int highlight);


    void handleSelection(int selection);

public:
    Menu();

    ~Menu();

    void openMenu();
    // void chooseRepo();
};

#endif