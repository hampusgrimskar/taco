#ifndef UTILS
#define UTILS

#include <iomanip>
#include <sstream>

#define NULL_CHARACTER '\0'

inline std::string generateId(const char* input)
{
    std::stringstream output_stream;
    for (int i = 0; input[i] != NULL_CHARACTER; i++)
    {
        // specify that each element in the stream should be represented with 2
        // hexadecimal characters. If the number only requires one character
        // then a leading '0' is inserted before
        output_stream << std::setw(2) << std::setfill('0') << std::hex << int(input[i]);
    }
    return output_stream.str();
}

#endif