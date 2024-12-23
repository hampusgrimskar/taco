#ifndef SESSION
#define SESSION

#include <string>
#include <iostream>
#include "ScreenCommand.h"
#include "Utils.h"

class Session 
{
private:
    const char* mySessionName;
    const char* myRepositoryName;
    std::string mySessionId;

    bool myIsRunning;

    void executeCommand(ScreenCommand commandType);

    void handleError(const std::exception& e);

public:
    Session(const char* session_name, const char* repository_name);

    ~Session();

    void attach();

    void detach();
};
#endif