We have a list of websites, and we want to check which ones are online and hwich are offline.
Doint this one-by-one is slow. So we will use concurrancy and mutithreading.

1. Define a function that checks a url. uses http.
    - this function needs its outputs to be in channels.
2. Make a function to gather all the info from different branches of that process that is for each url.
    - this function will get all the routines and the waits untill all of them are done?
