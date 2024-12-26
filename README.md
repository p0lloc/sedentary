# sedentary
Simple program for reminding myself to take regular breaks from the computer.  
Meant to be used in an X11 environment with libnotify.  

Usage: `./sedentary <work duration in seconds> <break duration in seconds>`

The program listens to keyboard/mouse events, and only starts a break if no activity was recorded for 30 seconds.  
When you come back from a break, simply use the keyboard/mouse to start a new work period.
