### cu-class-scheduler

automatic scheduler for CU boulder students

### Usage
```
 -break_time string
        time reserved for a break, or lunch. 24-hr format. Ex. 10:00-11:00
  -exclude_teacher string
        list of teachers to exclude, comma-separated, spaces after commas. Ex. E.Musk, K.Kardashian
  -include_teacher string
        list of teachers to include, comma-separated, spaces after commas. Ex. E.Musk, K.Kardashian
  -infile string
        input file with each class on a new line. (default "input.txt")
  -max_time string
        latest time for any class. 24-hr format. Ex. 20:00
  -mid_day string
        time of the day to prioritize classes from. 24-hr format. Ex. 12:00
  -min_buffer int
        Minimum time between classes, in minutes. Default is 0
  -min_time string
        earliest time for any class. 24-hr format. Ex. 8:00.
  -outfile string
        output file with a list of new schedules (default "output.txt")
  -rank_by_mid_day
        whether to prioritize classes from around the time specified by --mid_day (default is 12:00) (default true)
  -rank_by_teacher
        whether to rank results by included/excluded teachers. true/false. For this to work, you must specify lists of teachers to include/exclude using --include_teacher or --exclude_teacher.
  -topk int
        list of possible schedules to return, ranked by fit. Default is 10 (default 10)`
```

#### Features:

- [x] basic scheduling functionality
- [x] rank teachers using include_teacher=True
- [x] downrank teachers using exclude_teacher=True
- [x] implement start/end time, breaks 
- [x] implement rendering of calendars
- [ ] make calendar smaller by cropping
