#include "ftxui/component/event.hpp"
#include <unordered_map>

namespace std {
    template <>
    struct hash<ftxui::Event> {
        std::size_t operator()(const ftxui::Event& e) const {
            return std::hash<std::string>{}(e.input());
        }
    };
}

class FtxEvent
{
    public:
        enum EventType
        {
            CTRL_R,
            CTRL_C,
            RETURN,
            ALT_BACKSPACE,
            BACKSPACE,
            DELETE,
            UP,
            DOWN,
            RIGHT,
            LEFT,
            CHARACTER,
            OTHER
        };

    static EventType from(const ftxui::Event& event)
    {
        if (event.is_character()) return CHARACTER;

        static const std::unordered_map<ftxui::Event, EventType> mapping = {
            { ftxui::Event::CtrlC, CTRL_C },
            { ftxui::Event::Special({18}), CTRL_R },
            { ftxui::Event::Return, RETURN },
            { ftxui::Event::Special({27, 127}), ALT_BACKSPACE },
            { ftxui::Event::ArrowUp, UP },
            { ftxui::Event::ArrowDown, DOWN },
            { ftxui::Event::ArrowRight, RIGHT },
            { ftxui::Event::ArrowLeft, LEFT },
            { ftxui::Event::Backspace, BACKSPACE},
            { ftxui::Event::Delete, DELETE }
        };

        auto it = mapping.find(event);
        return it != mapping.end() ? it->second : OTHER;
    }
};
