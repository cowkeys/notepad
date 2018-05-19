# 拿带*号的pid
gops

# 查看状态
gops stats 9740

# 查看堆栈
gops stack 9740

# 输出浏览器
gops trace 9740

#
Commands:
    stack       Prints the stack trace.
    gc          Runs the garbage collector and blocks until successful.
    memstats    Prints the allocation and garbage collection stats.
    version     Prints the Go version used to build the program.
    stats       Prints the vital runtime stats.
    help        Prints this help text.

Profiling commands:
    trace       Runs the runtime tracer for 5 secs and launches "go tool trace".
    pprof-heap  Reads the heap profile and launches "go tool pprof".
    pprof-cpu   Reads the CPU profile and launches "go tool pprof".
