#include "Menu.h"

Menu::Menu()
{
    for (std::pair<std::string, std::string> repo : getReposFromConfigFile())
    {
        std::string session_name = repo.second != NO_ALIAS ? repo.second : repo.first;

        this->repositorySessionMap.push_back(
            std::make_tuple(session_name, repo.first, nullptr));
    }

    openMenu();
}

Menu::~Menu()
{
    for (auto repository_session : this->repositorySessionMap)
    {
        if (std::get<2>(repository_session) != nullptr) delete std::get<2>(repository_session);
    }
}

void Menu::printMenu(WINDOW *menu_win, int highlight)
{
	int x, y, i;	

	x = 2;
	y = 2;
	box(menu_win, 0, 0);
	for(i = 0; i < this->repositorySessionMap.size(); ++i)
	{	if(highlight == i) /* High light the present choice */
		{	wattron(menu_win, A_REVERSE); 
			mvwprintw(menu_win, y, x, "%s", std::get<0>(this->repositorySessionMap[i]).c_str());
			wattroff(menu_win, A_REVERSE);
		}
		else
			mvwprintw(menu_win, y, x, "%s", std::get<0>(this->repositorySessionMap[i]).c_str());
		++y;
	}
	wrefresh(menu_win);
}

void Menu::handleSelection(int selection)
{
    printf("You chose %s", std::get<0>(this->repositorySessionMap[selection]).c_str());

    if (std::get<2>(this->repositorySessionMap[selection]) == nullptr)
    {
        // If the repository is not currently associated with a sesion,
        // then create a new session
        const char* path = std::get<0>(this->repositorySessionMap[selection]).c_str();
        const char* alias = std::get<1>(this->repositorySessionMap[selection]).c_str();
        Session* session = new Session(path, alias);
        std::get<2>(this->repositorySessionMap[selection]) = session;

        // Attach to the new session
        // session->attach();
    }
    else
    {
        // Otherwise attach to the already existing session
        std::get<2>(this->repositorySessionMap[selection])->attach();        
    }
}

void Menu::openMenu()
{
    WINDOW *menu_win;
	int highlight = 0;
	int choice = -1;
	int c;

	initscr();
	start_color();
	clear();
	noecho();
	cbreak();	/* Line buffering disabled. pass on everything */
	int startx = (80 - 30) / 2;
	int starty = (24 - 10) / 2;
		
	curs_set(false);
	menu_win = newwin(9, 50, 7, 2);
	keypad(menu_win, TRUE);
	// printTitle();
	refresh();
	printMenu(menu_win, highlight);
	while(1)
	{	c = wgetch(menu_win);
		switch(c)
		{	case 107:
			case 259:
				if(highlight == 0)
					highlight = this->repositorySessionMap.size() - 1;
				else
					--highlight;
				break;
			case 106:
			case 258:
				if(highlight == this->repositorySessionMap.size() - 1)
					highlight = 0;
				else 
					++highlight;
				break;
			case 10:
				choice = highlight;
				break;
			case 27:
				nodelay(menu_win, true);
				endwin();
				break;
			default:
				refresh();
				break;
		}
		printMenu(menu_win, highlight);
		if(choice != -1)	/* User did a choice come out of the infinite loop */
			break;
	}	
	clrtoeol();
	refresh();
	endwin();

    // Attach to the session of the selected repository
    handleSelection(choice);
}