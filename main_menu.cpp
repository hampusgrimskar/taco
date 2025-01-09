#include <stdio.h>
#include <ncurses.h>

#define WIDTH 30
#define HEIGHT 10 

const char* TACO_TITLE = R"(
     ______   ______    ______    ______    
    /\__  _\ /\  __ \  /\  ___\  /\  __ \   
    \/_/\ \/ \ \  __ \ \ \ \____ \ \ \/\ \  
       \ \_\  \ \_\ \_\ \ \_____\ \ \_____\ 
        \/_/   \/_/\/_/  \/_____/  \/_____/ 
                                          
)";

void printTitle() {
	init_pair(1, COLOR_WHITE, COLOR_BLACK);

	char line[128];
    
    // Loop through the multiline string and print each line
    while (*TACO_TITLE) {
        int i = 0;
        while (*TACO_TITLE && *TACO_TITLE != '\n') {
            line[i++] = *TACO_TITLE++;
        }
        line[i] = '\0';  // Null-terminate the line

		attron(COLOR_PAIR(1));
        printw("%s\n", line);  // Print the current line
		attroff(COLOR_PAIR(1));

        if (*TACO_TITLE == '\n') {
            TACO_TITLE++;  // Skip the newline character
        }
    }
}

int startx = 0;
int starty = 0;

char *choices[] = { 
			"Choice 1",
			"Choice 2",
			"Choice 3",
			"Choice 4",
			"Exit",
		  };
int n_choices = sizeof(choices) / sizeof(char *);
void print_menu(WINDOW *menu_win, int highlight);

int main()
{	WINDOW *menu_win;
	int highlight = 1;
	int choice = 0;
	int c;

	initscr();
	start_color();
	clear();
	noecho();
	cbreak();	/* Line buffering disabled. pass on everything */
	startx = (80 - WIDTH) / 2;
	starty = (24 - HEIGHT) / 2;
		
	curs_set(false);
	menu_win = newwin(9, 50, 7, 2);
	keypad(menu_win, TRUE);
	printTitle();
	// mvprintw(0, 0, "Use arrow keys to go up and down, Press enter to select a choice");
	refresh();
	print_menu(menu_win, highlight);
	while(1)
	{	c = wgetch(menu_win);
		switch(c)
		{	case 107:
			case 259:
				if(highlight == 1)
					highlight = n_choices;
				else
					--highlight;
				break;
			case 106:
			case 258:
				if(highlight == n_choices)
					highlight = 1;
				else 
					++highlight;
				break;
			case 10:
				choice = highlight;
				break;
			default:
				mvprintw(24, 0, "Charcter pressed is = %3d Hopefully it can be printed as '%c'", c, c);
				refresh();
				break;
		}
		print_menu(menu_win, highlight);
		if(choice != 0)	/* User did a choice come out of the infinite loop */
			break;
	}	
	clrtoeol();
	refresh();
	endwin();
	printf("You chose %d", choice);
	return 0;
}


void print_menu(WINDOW *menu_win, int highlight)
{
	int x, y, i;	

	x = 2;
	y = 2;
	box(menu_win, 0, 0);
	for(i = 0; i < n_choices; ++i)
	{	if(highlight == i + 1) /* High light the present choice */
		{	wattron(menu_win, A_REVERSE); 
			mvwprintw(menu_win, y, x, "%s", choices[i]);
			wattroff(menu_win, A_REVERSE);
		}
		else
			mvwprintw(menu_win, y, x, "%s", choices[i]);
		++y;
	}
	wrefresh(menu_win);
}