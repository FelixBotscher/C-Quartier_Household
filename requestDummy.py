# to start this server run following command in the folder of the file (terminal): uvicorn requestDummy:app --reload 
import random
from fastapi import FastAPI
app = FastAPI()

randomVal = random.uniform(0.0,3000.0)
reply = "{\"version\":\"0.8.1\",\"generator\":\"vzlogger\",\"data\":[{\"uuid\":\"094fe804-1b2e-11ed-861d-0242ac120002\",\"last\":1651827462643,\"interval\":-1,\"protocol\":\"sml\",\"tuples\":[[1651827462643,215505.3]]},{\"uuid\":\"f924b3c8-199e-11ed-861d-0242ac120002\",\"last\":1651827462643,\"interval\":-1,\"protocol\":\"sml\",\"tuples\":[[1651827462643,"+ str(round(randomVal,2)) +"]]}]}"
workingReply = {"version":"0.8.1","generator":"vzlogger","data":[{"uuid":"094fe804-1b2e-11ed-861d-0242ac120002","last":1651827462643,"interval":-1,"protocol":"sml","tuples":[[1651827462643,215505.3]]},{"uuid":"f924b3c8-199e-11ed-861d-0242ac120002","last":1651827462643,"interval":-1,"protocol":"sml","tuples":[[1651827462643,round(randomVal,2)]]}]}

@app.get("/")
def replyMethod():
    return func(round(random.uniform(0,5000.0),2))

def func(var):
    return {"version":"0.8.1","generator":"vzlogger","data":[{"uuid":"094fe804-1b2e-11ed-861d-0242ac120002","last":1651827462643,"interval":-1,"protocol":"sml","tuples":[[1651827462643,215505.3]]},{"uuid":"f924b3c8-199e-11ed-861d-0242ac120002","last":1651827462643,"interval":-1,"protocol":"sml","tuples":[[1651827462643,var]]}]}
