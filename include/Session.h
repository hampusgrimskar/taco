#ifndef SESSION
#define SESSION

#include <string>
#include <iostream>
#include "ScreenCommand.h"

class Session 
{
private:
    const char* mySessionName;

    bool myIsRunning;

    void executeCommand(ScreenCommand commandType);

    void handleError(const std::exception& e);

public:
    Session(const char* session_name);

    ~Session();

    void attach();

    void detach();
};
#endif