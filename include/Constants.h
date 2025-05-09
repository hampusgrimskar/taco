#ifndef CONSTANTS
#define CONSTANTS

#include <string>

#define NULL_CHARACTER '\0'

inline const int MENU_SIZE = 17;
inline const int MENU_MARGIN = 5;
inline const int SESSION_NAME_AND_STATUS_GAP = 10;
inline const int SESSION_ACTIVE_PLACEHOLDER_LENGTH = 9;

inline const std::string FORWARD_SLASH = "/";
inline const std::string DOT = ".";
inline const std::string NO_ALIAS = "";
inline const std::string EMPTY_LINE = "";

inline std::string HOME_DIR = std::getenv("HOME");
inline std::string TACO_CONFIG_DIR = HOME_DIR + FORWARD_SLASH + DOT + "taco/";
inline std::string TACO_CONFIG_FILE = TACO_CONFIG_DIR + "repositories";

inline const std::string TACO_TITLE = R"(
     ______   ______    ______    ______
    /\__  _\ /\  __ \  /\  ___\  /\  __ \
    \/_/\ \/ \ \  __ \ \ \ \____ \ \ \/\ \
       \ \_\  \ \_\ \_\ \ \_____\ \ \_____\
        \/_/   \/_/\/_/  \/_____/  \/_____/

)";

#endif
