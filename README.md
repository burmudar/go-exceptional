Go-Exceptional: LogEvent Aggregator for Java Log files
===

The purpose of this program is to scan for ERROR events and find associated Caused By stack traces (typical of Java exceptions). An Error LogEvent is associated with a Caused By snippet
Both of these items are loaded into a DB. 


Using the ERROR events that are loaded in the DB. The standard deviation is calculated based on a day. 
If while the program is watching the current log file the standard deviation (in terms of particular errors) is exceeded an alarm is sounded. 
The Alarm for now, is a call to an external server which will notify whomever it is configured to notify

What is working
===

*   Parsing of log files into log events [complete]
*   Filter logevents into error events with Caused by [complete]
*   Loading SQLIte database with error events [complete]
*   Calculate daily summaries on events [complete]
*   Statistics on exceptions [complete]
*   Sound alarm based on standard deviation of exceptions [complete]
