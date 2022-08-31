from flask import Flask
from flask import request

app = Flask(__name__)

@app.post("/schedule")
def scheduler_run():
    # we dont return anything except a confirmation that the scheduler ran
    # index.html will just pull images from schedule folder

    assert type(request.form['classes']) == list[Str]
    time_between_classess = request.form('time_between_classes')
    min_time_string = request.form('min_time_string')
    max_time_string = request.form('max_time_string')
    break_time_string = request.form('break_time_string')
    infile = request.form('infile')
    outfile = request.form('outfile')


    return "<p>Hello, World!</p>"