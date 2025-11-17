#include "FtxMenu.h"

FtxMenu::FtxMenu()
{
    this->longest_title = 0;
    this->selected_item = 0;
	updateRepositorySessions();
}

FtxMenu::~FtxMenu()
{
    deleteSessions();
}

void FtxMenu::updateRepositorySessions()
{
	for (std::pair<std::string, std::string> repo : getReposFromConfigFile())
	{
		std::string session_name = repo.second != NO_ALIAS ? repo.second : repo.first;

		// Skip empty lines in config file
		if (session_name == EMPTY_LINE) continue;

		FtxMenu::RepositorySession repository_session(session_name, repo.first);

		if (std::find(this->repositorySessions.begin(),
			this->repositorySessions.end(),
			repository_session) == this->repositorySessions.end())
			{
				this->repositorySessions.push_back(repository_session);
			}

		updateLongestTitle(session_name);
	}
}

void FtxMenu::attachSession(int selection)
{
    FtxMenu::RepositorySession* selected_repository = &this->repositorySessions[selection];

        if (selected_repository->session == nullptr)
        {
            // If the repository is not currently associated with a session,
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

void FtxMenu::sortRepositorySessions()
{
    RepositorySession selected_repo = repositorySessions[selected_item];

	auto sortByStatus = [](const RepositorySession& rs1, const RepositorySession& rs2) {
		return (rs1.is_active && !rs2.is_active);
	};
	std::stable_sort(this->repositorySessions.begin(), this->repositorySessions.end(), sortByStatus);

    auto it = std::find(repositorySessions.begin(), repositorySessions.end(), selected_repo);
	selected_item = std::distance(repositorySessions.begin(), it);
}

std::vector<std::string> FtxMenu::getMenuEntries()
{
    sortRepositorySessions();

    std::vector<std::string> menu_entries;
    std::transform(
        repositorySessions.begin(),
        repositorySessions.end(),
        std::back_inserter(menu_entries),
        [this](const RepositorySession& rs) {
            return getSessionNameWithStatus(rs);
    });
    return menu_entries;
}

void FtxMenu::show()
{
    using namespace ftxui;

    auto screen = ScreenInteractive::Fullscreen();

    std::vector<std::string> menu_entries = getMenuEntries();

    screen.TrackMouse(false);
    
    bool should_attach = false;
    bool should_exit = false;
    
    auto menu_option = MenuOption();
    menu_option.focused_entry = selected_item;
    menu_option.on_enter = [&]() {
        should_attach = true;
        screen.Post([&] { screen.Exit(); });
    };

    auto menu = Menu(&menu_entries, &selected_item, menu_option);
    
    auto component = CatchEvent(menu, [&](Event event) {
        if (event == Event::CtrlC) {
            should_exit = true;
            screen.Post([&] { screen.Exit(); });
            return true;
        }
        // This is not pretty... But it works
        if (event == Event::ArrowUp && selected_item == 0) {
            // Simulate pressing down arrow multiple times to reach the end
            for (int i = 0; i < menu_entries.size() - 1; i++) {
                menu->OnEvent(Event::ArrowDown);
            }
            return true;
        }
        if (event == Event::ArrowDown && selected_item == menu_entries.size() - 1) {
            // Simulate pressing up arrow multiple times to reach the beginning
            for (int i = 0; i < menu_entries.size() - 1; i++) {
                menu->OnEvent(Event::ArrowUp);
            }
            return true;
        }
        return false;
    });

    auto renderer = Renderer(component, [&] {
        Element menu_element = menu->Render() | vscroll_indicator | frame | border;
        return vbox({
            filler(),
            hbox({filler(), menu_element, filler()}),
            filler()
        });
    });

    // Menu event loop
    screen.Loop(renderer);
    
    if (should_exit)
    {
        deleteSessions();
        exit(0);
    }
    
    if (should_attach)
    {
        attachSession(selected_item);
    }
}
