#include "Menu.h"

Menu::Menu()
{
	for (std::pair<std::string, std::string> repo : getReposFromConfigFile())
	{
		std::string session_name = repo.second != NO_ALIAS ? repo.second : repo.first;

		Menu::RepositorySession repository_session(session_name, repo.first);
		this->repositorySessions.push_back(repository_session);
	}
}

Menu::~Menu()
{
	// Delete Session objects
	for (auto repository_session : this->repositorySessions)
	{
		if (repository_session.session != nullptr)
		{
			delete repository_session.session;
		}
	}
	// Stop Curses window
	endwin();
}

void Menu::printTitle()
{
	printw("%s\n", TACO_TITLE.c_str());
}

void Menu::printMenu(WINDOW *menu_win, int highlight)
{
	int x, y, i;

	x = 2;
	y = 2;
	box(menu_win, 0, 0);
	for (i = 0; i < this->repositorySessions.size(); ++i)
	{
		if (highlight == i) /* High light the present choice */
		{
			wattron(menu_win, A_REVERSE);
			mvwprintw(menu_win, y, x, "%s", this->repositorySessions[i].getSessionNameWithStatus().c_str());
			wattroff(menu_win, A_REVERSE);
		}
		else
			mvwprintw(menu_win, y, x, "%s", this->repositorySessions[i].getSessionNameWithStatus().c_str());
		++y;
	}
	wrefresh(menu_win);
}

void Menu::handleSelection(int selection)
{
	Menu::RepositorySession& selected_repository = this->repositorySessions[selection];

	if (this->repositorySessions[selection].session == nullptr)
	{
		// If the repository is not currently associated with a sesion,
		// then create a new session and mark it as active
		const char *alias = selected_repository.session_name.c_str();
		const char *path = selected_repository.path.c_str();

		Session *session = new Session(alias, path);

		selected_repository.session = session;
		selected_repository.is_active = true;
	}
	else
	{
		// Otherwise attach to the already existing session
		selected_repository.session->attach();
	}
}

/*
This method is ugly, needs cleanup in the future
*/
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
	cbreak(); // Line buffering disabled. pass on everything
	int startx = (80 - 30) / 2;
	int starty = (24 - 10) / 2;

	curs_set(false);
	menu_win = newwin(9, 50, 7, 2);
	keypad(menu_win, TRUE);
	printTitle();
	refresh();
	printMenu(menu_win, highlight);
	while (1)
	{
		c = wgetch(menu_win);
		switch (c)
		{
		case 107:
		case 259:
			if (highlight == 0)
				highlight = this->repositorySessions.size() - 1;
			else
				--highlight;
			break;
		case 106:
		case 258:
			if (highlight == this->repositorySessions.size() - 1)
				highlight = 0;
			else
				++highlight;
			break;
		case 10:
			choice = highlight;
			break;
		case 27:
			nodelay(menu_win, true);
			break;
		default:
			refresh();
			break;
		}
		printMenu(menu_win, highlight);
		if (choice != -1) // User did a choice come out of the infinite loop
			break;
	}
	clrtoeol();
	refresh();
	endwin();

	// Attach to the session of the selected repository
	handleSelection(choice);
}
