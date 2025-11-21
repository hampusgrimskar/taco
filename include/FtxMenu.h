#ifndef FTXMENU
#define FTXMENU

#include "Session.h"
#include <stdlib.h>
#include <functional>
#include <memory>
#include <iostream>
#include <string>
#include <vector>
#include "ftxui/component/screen_interactive.hpp"
#include <locale>
#include <ftxui/dom/elements.hpp>
#include <ftxui/screen/screen.hpp>
#include "ftxui/dom/node.hpp"
#include "ftxui/screen/color.hpp"
#include <ftxui/dom/table.hpp>
#include <ftxui/screen/screen.hpp>
#include "ftxui/component/captured_mouse.hpp"
#include "ftxui/component/component.hpp"
#include "ftxui/component/component_base.hpp"
#include "ftxui/component/component_options.hpp"
#include "ftxui/component/screen_interactive.hpp"
#include "ftxui/dom/elements.hpp"

class FtxMenu
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

        bool operator==(const RepositorySession& other)
        {
            return (this->session_name == other.session_name)
                && (this->path == other.path);
        }
    };

    int longest_title;
    int selected_item;
    std::vector<FtxMenu::RepositorySession> repositorySessions;
    FtxMenu::RepositorySession* selected_rs = nullptr;

    void updateLongestTitle(const std::string& session_name)
    {
        if (session_name.length() > this->longest_title)
        {
            longest_title = session_name.length();
        }
    }

    std::string getSessionNameWithStatus(const FtxMenu::RepositorySession& repositorySession)
    {
        int number_of_spaces = longest_title - repositorySession.session_name.length() + SESSION_NAME_AND_STATUS_GAP;
        std::string spaces = std::string(number_of_spaces, ' ');
        return (repositorySession.session_name + (repositorySession.is_active ? spaces + " (Active)" : spaces + "         "));
    }

    void deleteSessions()
    {
        // Delete Session objects
        for (auto repository_session : this->repositorySessions)
        {
            if (repository_session.session != nullptr)
            {
                delete repository_session.session;
            }
        }
    }

    void updateRepositorySessions();

    void sortRepositorySessions();

    void attachSession(FtxMenu::RepositorySession* rs);

    std::vector<std::string> getMenuEntries();

    FtxMenu::RepositorySession* find_selected_repository_session(int selection, std::vector<std::string> menu_entries);

    ftxui::Component ModalComponent(std::function<void()> hide_modal);

public:
    FtxMenu();
    ~FtxMenu();
    void show();
};

#endif