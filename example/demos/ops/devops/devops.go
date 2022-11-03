package devops

import (
	"bufio"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/hexasoftware/flow/registry"
)

// New create a new devops Registry
func New() *registry.R {

	r := registry.New()

	r.Add(
		dockerNew,
		setWriter,
		dockerTest,
	).Tags("build")

	return r
}

//////////////////////
// DevOps SIM
////////

// DockerHandler example
type DockerHandler struct {
	out io.Writer
}

func dockerNew() DockerHandler {
	return DockerHandler{}
}

func setWriter(out io.Writer, o DockerHandler) DockerHandler {
	o.out = out
	return o
}

//
func dockerTest(handler DockerHandler, imageName string) DockerHandler {
	sampleData := `
	make: Entering directory '/home/stdio/coding/Projects/Flow'
make -C go test
make[1]: Entering directory '/home/stdio/coding/Projects/Flow/go'
gocov test -race ./src/... | gocov report
ok  	flow	1.020s	coverage: 84.1% of statements
?   	flow/cmd/buildops	[no test files]
?   	flow/cmd/demo1	[no test files]
?   	flow/flowserver	[no test files]
?   	flow/flowserver/flowmsg	[no test files]
?   	flow/internal/assert	[no test files]
ok  	flow/registry	1.011s	coverage: 100.0% of statements

flow/flow.go		 Flow.String			 100.00% (11/11)
flow/utils.go		 RandString			 100.00% (10/10)
flow/flow.go		 @290:21			 100.00% (8/8)
flow/flow.go		 Flow.MarshalJSON		 100.00% (8/8)
flow/flow.go		 Flow.Const			 100.00% (7/7)
flow/flow.go		 Flow.Var			 100.00% (6/6)
flow/flow.go		 @315:21			 100.00% (6/6)
flow/hook.go		 Hooks.Attach			 100.00% (3/3)
flow/flow.go		 Flow.Run			 100.00% (2/2)
flow/flow.go		 Flow.SetRegistry		 100.00% (2/2)
flow/operation.go	 @107:12			 100.00% (2/2)
flow/operation.go	 operation.ID			 100.00% (1/1)
flow/operation.go	 opFunc				 100.00% (1/1)
flow/operation.go	 opVar				 100.00% (1/1)
flow/operation.go	 opConst			 100.00% (1/1)
flow/flow.go		 Flow.In			 100.00% (1/1)
flow/operation.go	 @220:12			 100.00% (1/1)
flow/flow.go		 Flow.Res			 100.00% (1/1)
flow/operation.go	 @221:12			 100.00% (1/1)
flow/operation.go	 opIn				 100.00% (1/1)
flow/flow.go		 Flow.Hook			 100.00% (1/1)
flow/operation.go	 operation.Set			 100.00% (1/1)
flow/hook.go		 Hooks.wait			 100.00% (1/1)
flow/flow.go		 Flow.SetIDGen			 100.00% (1/1)
flow/hook.go		 Hooks.finish			 100.00% (1/1)
flow/operation.go	 operation.processWithCtx	 100.00% (1/1)
flow/flow.go		 @49:13				 100.00% (1/1)
flow/operation.go	 newOpCtx			 100.00% (1/1)
flow/operation.go	 operation.Process		 100.00% (1/1)
flow/hook.go		 Hooks.start			 100.00% (1/1)
flow/flow.go		 New				 100.00% (1/1)
flow/flow.go		 Flow.DefOp			 92.31% (12/13)
flow/flow.go		 Flow.Op			 88.89% (16/18)
flow/flow.go		 @240:21			 84.21% (16/19)
flow/flow.go		 Flow.run			 84.21% (16/19)
flow/operation.go	 @166:8				 80.00% (4/5)
flow/hook.go		 Hooks.Trigger			 78.57% (11/14)
flow/flow.go		 Flow.Analyse			 75.00% (3/4)
flow/flow.go		 Flow.getOp			 75.00% (3/4)
flow/flow.go		 Flow.Must			 66.67% (2/3)
flow/operation.go	 @93:12				 66.67% (2/3)
flow/operation.go	 @121:12			 65.22% (30/46)
flow/operation.go	 @123:10			 33.33% (1/3)
flow/hook.go		 Hooks.error			 0.00% (0/1)
flow/operation.go	 opNil				 0.00% (0/1)
flow/operation.go	 @229:12			 0.00% (0/1)
flow/operation.go	 dumbSet			 0.00% (0/0)
flow			 ------------------------	 84.10% (201/239)

flow/registry/entry.go		 NewEntry		 100.00% (18/18)
flow/registry/registry.go	 R.Get			 100.00% (12/12)
flow/registry/registry.go	 R.Add			 100.00% (8/8)
flow/registry/entry.go		 Entry.DescInputs	 100.00% (8/8)
flow/registry/batch.go		 Batch			 100.00% (6/6)
flow/registry/registry.go	 R.Register		 100.00% (5/5)
flow/registry/entry.go		 Entry.Extra		 100.00% (4/4)
flow/registry/registry.go	 R.Entry		 100.00% (4/4)
flow/registry/registry.go	 R.Clone		 100.00% (4/4)
flow/registry/entry.go		 Entry.Tags		 100.00% (4/4)
flow/registry/entry.go		 Entry.DescOutput	 100.00% (4/4)
flow/registry/registry.go	 R.Descriptions		 100.00% (4/4)
flow/registry/batch.go		 EntryBatch.Extra	 100.00% (3/3)
flow/registry/batch.go		 EntryBatch.DescOutput	 100.00% (3/3)
flow/registry/batch.go		 EntryBatch.DescInputs	 100.00% (3/3)
flow/registry/batch.go		 EntryBatch.Tags	 100.00% (3/3)
flow/registry/entry.go		 Entry.Err		 100.00% (1/1)
flow/registry/registry.go	 New			 100.00% (1/1)
flow/registry			 ---------------------	 100.00% (95/95)

Total Coverage: 88.62% (296/334)
make[1]: Leaving directory '/home/stdio/coding/Projects/Flow/go'
make: Leaving directory '/home/stdio/coding/Projects/Flow'`

	scanner := bufio.NewScanner(strings.NewReader(sampleData))

	for scanner.Scan() {
		time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
		if handler.out != nil {
			handler.out.Write([]byte(scanner.Text()))
		}
	}

	return handler
}
