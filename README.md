Go-Exceptional: LogEvent Aggregator for Java Log files
===

The purpose of this program is to scan for ERROR events and find associated Caused By stack traces (typical of Java exceptions). An Error LogEvent is associated with a Caused By snippet
Both of these items are loaded into a DB. 


Using the ERROR events that are loaded in the DB. The standard deviation is calculated based on a day. 
If while the program is watching the current log file the standard deviation (in terms of particular errors) is exceeded an alarm is sounded. 
The Alarm for now, is a call to an external server which will notify whomever it is configured to notify

What is working
===

Not much :P

*   Parsing of log files into log events [complete]
*   Loading SQLIte database with logevents [incomplete]
*   Sound alarm based on standard deviation [incomplete]
