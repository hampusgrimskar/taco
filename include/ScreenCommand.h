#ifndef SCREENCOMMAND
#define SCREENCOMMAND

namespace ScreenConstants
{
    inline const char *CREATE_SESSION_COMMAND = "screen -S %s bash -c \"screen -S %s -X bindkey ^d detach && screen -S %s -X bindkey ^[ detach && cd %s && exec bash\"";
    inline const char *ATTACH_TO_SESSION_COMMAND = "screen -r %s";
    inline const char *DETACH_SESSION_COMMAND = "screen -d %s";
    inline const char *TERMINATE_SESSION_COMMAND = "screen -X -S %s quit";
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
