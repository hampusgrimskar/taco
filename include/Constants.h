#ifndef CONSTANTS
#define CONSTANTS

#include <string>

#define NULL_CHARACTER '\0'

inline const std::string FORWARD_SLASH = "/";
inline const std::string DOT = ".";
inline const std::string NO_ALIAS = "";

inline std::string HOME_DIR = std::getenv("HOME");
inline std::string TACO_CONFIG_DIR = HOME_DIR + FORWARD_SLASH + DOT + "taco/";
inline std::string TACO_CONFIG_FILE = TACO_CONFIG_DIR + "repositories";

// const char* TACO_TITLE = R"(
//      ______   ______    ______    ______    
//     /\__  _\ /\  __ \  /\  ___\  /\  __ \   
//     \/_/\ \/ \ \  __ \ \ \ \____ \ \ \/\ \  
//        \ \_\  \ \_\ \_\ \ \_____\ \ \_____\ 
//         \/_/   \/_/\/_/  \/_____/  \/_____/ 
                                          
// )";

#endif