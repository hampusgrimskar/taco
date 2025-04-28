#include "Menu.h"
#include <algorithm>

Menu::Menu()
{
	this->longest_title = 0;
	updateRepositorySessions();
	slidingWindow = { repositorySessions.begin(), repositorySessions.begin() + MENU_SIZE };
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
	RepositorySession highlighted_repo = repositorySessions[highlight];

	auto sortByStatus = [](const RepositorySession& rs1, const RepositorySession& rs2) {
		return (rs1.is_active && !rs2.is_active);
	};
	std::stable_sort(this->repositorySessions.begin(), this->repositorySessions.end(), sortByStatus);

	auto it = std::find(repositorySessions.begin(), repositorySessions.end(), highlighted_repo);
	highlight = std::distance(repositorySessions.begin(), it);
}

void Menu::slideMenuDown()
{
	
	if (highlight > MENU_SIZE - 1 + slidingOffset)
	{
		slidingOffset++;
	}

	// Handle wrap around from bottom to top
	if (highlight == 0)
	{
		slidingOffset = 0;
	}

	slidingWindow = {
		repositorySessions.begin() + slidingOffset,
		repositorySessions.begin() + MENU_SIZE + slidingOffset };
}

void Menu::slideMenuUp()
{
	
	if (highlight < slidingOffset)
	{
		slidingOffset--;
	}

	// Handle wrap around from top to bottom
	if (highlight == repositorySessions.size() - 1)
	{
		slidingOffset = repositorySessions.size() - MENU_SIZE;
	}

	slidingWindow = {
		repositorySessions.begin() + slidingOffset,
		repositorySessions.begin() + MENU_SIZE + slidingOffset };
}

void Menu::printMenu(WINDOW *menu_win)
{
	sortRepositorySessions(); // Put active sessions on top

	int x, y;

	x = 2;
	y = 2;
	box(menu_win, 0, 0);
	mvwprintw(menu_win, 0, 2, "%s", "MY REPOSITORIES");
	for (int i = 0; i < slidingWindow.size(); ++i)
	{
		if (highlight == slidingOffset + i) // Highlight the present choice
		{
			wattron(menu_win, A_REVERSE);
			mvwprintw(menu_win, y, x, "%s", getSessionNameWithStatus(slidingWindow[i]).c_str());
			wattroff(menu_win, A_REVERSE);
		}
		else
		{
			mvwprintw(menu_win, y, x, "%s", getSessionNameWithStatus(slidingWindow[i]).c_str());
		}
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
}

void Menu::removeRepositorySession(const std::string session_name)
{
	auto it = std::find_if(this->repositorySessions.begin(), this->repositorySessions.end(),
	[session_name](const RepositorySession& rs) {
		return rs.session_name == session_name;
	});

	if (it != this->repositorySessions.end())
	{
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
	int choice = -1;
	int c;

	initscr();
	int terminal_width, terminal_height;
	getmaxyx(stdscr, terminal_height, terminal_width);
	clear();
	noecho();
	cbreak(); // Line buffering disabled. pass on everything

	int menu_hight = this->repositorySessions.size() + 3;
	int menu_width = longest_title + SESSION_ACTIVE_PLACEHOLDER_LENGTH + MENU_MARGIN;

	curs_set(false);
	menu_win = newwin(
		menu_hight,
		menu_width,
		(terminal_height / 2) - (menu_hight / 2),
		(terminal_width / 2) - (menu_width / 2)
	);

	keypad(menu_win, TRUE);
	refresh();

	printMenu(menu_win);

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
				slideMenuUp();
			break;
		case 106:
		case 258:
			if (highlight == this->repositorySessions.size() - 1)
				highlight = 0;
			else
				++highlight;
				slideMenuDown();
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
		printMenu(menu_win);
		if (choice != -1) // User did a choice come out of the infinite loop
			break;
	}
	clrtoeol();
	refresh();
	endwin();

	// Attach to the session of the selected repository
	handleSelection(choice);
}
