#ifndef SCREENCOMMAND
#define SCREENCOMMAND

namespace ScreenConstants
{
    inline const char *CREATE_SESSION_COMMAND =
        "tmux new-session -s %s \\\n"
        " -c %s \\\n"
        "\"tmux bind-key -n Escape run 'if \\\n"
        " [ #{pane_current_command} != vi ] && \\\n"
        " [ #{pane_current_command} != vim ] && \\\n"
        " [ #{pane_current_command} != nvim ] && \\\n"
        " [ #{pane_current_command} != k9s ] && \\\n"
        " [ #{pane_current_command} != git ]; \\\n"
        " then tmux detach; \\\n"
        " else tmux send-keys Escape; fi'; \\\n"
        " tmux set-option -g escape-time 0; \\\n"
        " tmux set -g mouse on; \\\n"
        " bash\"\n";

    inline const char *ATTACH_TO_SESSION_COMMAND = "tmux attach -t %s";
    inline const char *DETACH_SESSION_COMMAND = "tmux detach -s %s";
    inline const char *TERMINATE_SESSION_COMMAND = "tmux kill-session -t %s";
}

enum ScreenCommand
{
    CREATE_SESSION,
    ATTACH_TO_SESSION,
    DETACH_SESSION,
    TERMINATE_SESSION
};

inline char *getCommand(ScreenCommand command_type)
{
    switch (command_type)
    {
    case CREATE_SESSION:
        return (char *)ScreenConstants::CREATE_SESSION_COMMAND;

    case ATTACH_TO_SESSION:
        return (char *)ScreenConstants::ATTACH_TO_SESSION_COMMAND;

    case DETACH_SESSION:
        return (char *)ScreenConstants::DETACH_SESSION_COMMAND;

    case TERMINATE_SESSION:
        return (char *)ScreenConstants::TERMINATE_SESSION_COMMAND;

    default:
        return NULL;
    }
}

#endif
