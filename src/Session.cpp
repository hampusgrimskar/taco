#include "Session.h"

Session::Session(const char *session_name, const char *repository_name) : mySessionName(session_name),
                                                                          myRepositoryName(repository_name),
                                                                          mySessionId(generateId())
{
    // Start new terminal session
    try
    {
        executeCommand(CREATE_SESSION);
    }
    catch (const std::exception &e)
    {
        myIsRunning = false;
        delete this;
        handleError(e);
        return;
    }

    // Set to true if a new terminal session was created succesfully
    myIsRunning = true;
}

Session::~Session()
{
    if (!myIsRunning)
        return; // Don't cancel terminal session if it was never started
    try
    {
        executeCommand(TERMINATE_SESSION);
    }
    catch (const std::exception &e)
    {
        handleError(e);
    }
}

void Session::attach()
{
    try
    {
        executeCommand(ATTACH_TO_SESSION);
    }
    catch (const std::exception &e)
    {
        handleError(e);
    }
}

void Session::detach()
{
    try
    {
        executeCommand(DETACH_SESSION);
    }
    catch (const std::exception &e)
    {
        handleError(e);
    }
}

void Session::executeCommand(ScreenCommand command_type)
{
    char f_command[1024];

    if (command_type == CREATE_SESSION)
    {
        sprintf(f_command, getCommand(command_type), mySessionId.c_str(), myRepositoryName);
    }
    else
    {
        sprintf(f_command, getCommand(command_type), mySessionId.c_str());
    }

    system(f_command);
}

void Session::handleError(const std::exception &e)
{
    // Maybe this method should do other things later,
    // for now just print error
    std::cout << e.what() << std::endl;
}
