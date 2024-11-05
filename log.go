package main



type Logger struct{
    Id string
}


func (Logger)NewLogger()Logger{
    logger := Logger{Id:timestamp()}
    return logger
}

func (Logger)EnableLog(verbose bool)


var Logg Logger = Logger{}.NewLogger()

