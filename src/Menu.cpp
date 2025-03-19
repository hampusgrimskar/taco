#include "Menu.h"
#include <algorithm>

Menu::Menu()
{
	this->longest_title = 0;
	updateRepositorySessions();
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

void Menu::updateRepositorySessions()
{
	for (std::pair<std::string, std::string> repo : getReposFromConfigFile())
	{
		std::string session_name = repo.second != NO_ALIAS ? repo.second : repo.first;

		// Skip empty lines in config file
		if (session_name == EMPTY_LINE) continue;

		Menu::RepositorySession repository_session(session_name, repo.first);

		if (std::find(this->repositorySessions.begin(),
			this->repositorySessions.end(),
			repository_session) == this->repositorySessions.end())
			{
				this->repositorySessions.push_back(repository_session);
			}

		updateLongestTitle(session_name);
	}
}

void Menu::sortRepositorySessions()
{
	auto sortByStatus = [](const RepositorySession& rs1, const RepositorySession& rs2) {
		return (rs1.is_active && !rs2.is_active);
	};
	std::sort(this->repositorySessions.begin(), this->repositorySessions.end(), sortByStatus);
}

void Menu::printMenu(WINDOW *menu_win, int highlight)
{
	sortRepositorySessions(); // Put active sessions on top

	int x, y, i;

	x = 2;
	y = 2;
	box(menu_win, 0, 0);
	mvwprintw(menu_win, 0, 2, "%s", "MY REPOSITORIES");
	for (i = 0; i < this->repositorySessions.size(); ++i)
	{
		if (highlight == i) // Highlight the present choice
		{
			wattron(menu_win, A_REVERSE);
			mvwprintw(menu_win, y, x, "%s", getSessionNameWithStatus(this->repositorySessions[i]).c_str());
			wattroff(menu_win, A_REVERSE);
		}
		else
			mvwprintw(menu_win, y, x, "%s", getSessionNameWithStatus(this->repositorySessions[i]).c_str());
		++y;
	}
	wrefresh(menu_win);
}

void Menu::handleSelection(int selection)
{
	Menu::RepositorySession* selected_repository = &this->repositorySessions[selection];

	if (this->repositorySessions[selection].session == nullptr)
	{
		// If the repository is not currently associated with a sesion,
		// then create a new session and mark it as active
		const char *alias = selected_repository->session_name.c_str();
		const char *path = selected_repository->path.c_str();

		Session *session = new Session(alias, path);

		selected_repository->session = session;
		selected_repository->is_active = true;
	}
	else
	{
		// Otherwise attach to the already existing session
		selected_repository->session->attach();
	}
	std::cout << this->repositorySessions[selection].session << std::endl;
}

void Menu::removeRepositorySession(const std::string session_name)
{
	for (RepositorySession ses : this->repositorySessions)
	{
		std::cout << ses.session << std::endl;
	}
	auto it = std::find_if(this->repositorySessions.begin(), this->repositorySessions.end(),
	[session_name](const RepositorySession& rs) {
		return rs.session_name == session_name;
	});

	if (it != this->repositorySessions.end())
	{
		std::cout << it->session << std::endl;
		it->session->detach();
		it->session->~Session();
	}
}

void Menu::removeRepository(const cxxopts::ParseResult &result)
{
    std::string pwd = std::getenv("PWD");

    if (!isRepoInitialized(pwd)) return;

    std::ifstream file_in(TACO_CONFIG_FILE);
    std::ofstream temp(TACO_CONFIG_FILE + "-temp", std::ios::app);
    std::string line;
    std::unordered_map<std::string, std::string> repos;
    while (getline(file_in, line))
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

        if (path == pwd)
        {
            try
            {
                this->removeRepositorySession(alias != NO_ALIAS ? alias : path);
            }
            catch(const std::exception& e)
            {
                std::cerr << e.what() << '\n';
                file_in.close();
                temp.close();
                return;
            }
            
            line.replace(0, line.length(), EMPTY_LINE);
        }
        temp << line << std::endl;
    }
    file_in.close();
    temp.close();
    remove(TACO_CONFIG_FILE.c_str());
    rename((TACO_CONFIG_FILE + "-temp").c_str(), TACO_CONFIG_FILE.c_str());
}

/*
This method is ugly, needs cleanup in the future
*/
void Menu::openMenu()
{
	updateRepositorySessions();

	WINDOW *menu_win;
	int highlight = 0;
	int choice = -1;
	int c;

	initscr();
	int terminal_width, terminal_height;
	getmaxyx(stdscr, terminal_height, terminal_width);
	clear();
	noecho();
	cbreak(); // Line buffering disabled. pass on everything
	// int startx = (80 - 30) / 2;
	// int starty = (24 - 10) / 2;
	int menu_hight = this->repositorySessions.size() + 3;
	int menu_width = 50;

	curs_set(false);
	menu_win = newwin(
		menu_hight,
		menu_width,
		(terminal_height / 2) - (menu_hight / 2),
		(terminal_width / 2) - (menu_width / 2)
	);

	keypad(menu_win, TRUE);
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
		case 115:
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
