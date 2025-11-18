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

void FtxMenu::sortRepositorySessions()
{
    RepositorySession selected_repo = selected_rs != nullptr ? *selected_rs : repositorySessions[selected_item];

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

    std::vector<std::string> original_menu_entries = getMenuEntries();
    std::vector<std::string> menu_entries = original_menu_entries;

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
    
    std::string search_str;
    
    InputOption input_option;
    input_option.placeholder = "search...";
    Component search_input = Input(&search_str, input_option);

    auto component = CatchEvent(menu, [&](Event event) {
        if (event == Event::CtrlC) {
            should_exit = true;
            screen.Post([&] { screen.Exit(); });
            return true;
        }
        
        // Let search input handle text input, but menu handles navigation
        if (event.is_character() || event == Event::Backspace || event == Event::Delete) {
            return search_input->OnEvent(event);
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
            // Filter menu entries based on search str
            menu_entries = original_menu_entries;
            std::vector<std::string> filtered_entries;
            std::copy_if(
                menu_entries.begin(),
                menu_entries.end(),
                std::back_inserter(filtered_entries),
                [&search_str](const std::string& ref) {
                    // if session_name contains search_str
                    return ref.find(search_str) != std::string::npos;
                });
            menu_entries = filtered_entries;

        // Calculate width based on longest menu entry
        int max_width = 0;
        for (const auto& entry : menu_entries) {
            max_width = std::max(max_width, (int)entry.length());
        }
        max_width += 5; // Add padding
        
        Element menu_element = menu->Render() | vscroll_indicator | frame | border;
        Element search_element = hbox({
            // text("Search: "),
            text(search_str.empty() ? "search..." : search_str) | flex
        }) | frame | border | size(WIDTH, EQUAL, max_width);
        
        return vbox({
            filler(),
            hbox({filler(), menu_element, filler()}),
            hbox({filler(), search_element, filler()}),
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
        find_selected_repository_session(selected_item, menu_entries);
        attachSession(selected_rs);
    }
}

FtxMenu::RepositorySession* FtxMenu::find_selected_repository_session(int selection, std::vector<std::string> menu_entries)
{
    std::string entry = menu_entries[selected_item];

    for (RepositorySession& rs : repositorySessions)
    {
        if (getSessionNameWithStatus(rs) == entry)
        {
            selected_rs = &rs;
            break;
        }
    }

    return selected_rs;
}

void FtxMenu::attachSession(FtxMenu::RepositorySession* rs)
{
    if (rs->session == nullptr)
    {
        // If the repository is not currently associated with a session,
        // then create a new session and mark it as active
        const char *alias = rs->session_name.c_str();
        const char *path = rs->path.c_str();

        Session *session = new Session(alias, path);

        rs->session = session;
        rs->is_active = true;
    }
    else
    {
        // Otherwise attach to the already existing session
        rs->session->attach();
    }
}