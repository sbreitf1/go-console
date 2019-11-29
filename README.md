# Go Console

Util package for convenient terminal input and output including a command line environment with command history and fully customizable completion.

See section `Command Line Environment` for a sophisticated and highly customizable `bash`-like input environment. Section `Custom Completion Handlers` contains some details on how command completion works and can be customized.

## Basic Input & Output

This package includes the most common output methods known from the `fmt` package:

```golang
console.Print("foo", "bar")          // outputs "foo bar"
console.Printf("hello %s", "world")  // outputs "hello world"
console.Println("foobar")            // outputs "foobar\n"
console.Printlnf("foo%s", "bar")     // outputs "foobar\n")
```

For basic input you can use `ReadLine` and `ReadPassword` for hidden input:

```golang
line, err := console.ReadLine()
password, err := console.ReadPassword() // will hide input while typing
```

See `examples/basic-input` for an example application.

## Command Input

A more advanced input method is provided by `ReadCommand`. It reads and parses a command from input, respecting all escape characters and quoted phrases:

```golang
cmd, err := console.ReadCommand("prompt", nil) // prompt> {user input here}
// example user input: echo foo 'say "hello world"' "white space" escape\ sequence
// cmd[0] = "echo"
// cmd[1] = "foo"
// cmd[2] = "say \"hello world\""
// cmd[3] = "white space"
// cmd[4] = "escape sequence"
```

You can additionally pass handlers for command history (up and down arrow keys), aswell as completion (tab key). Consider using a `Command Line Environment` for command-based applications.

See `examples/read-command` for an example application.

## Command Line Environment

The most sophisticated input method is by instantiating a `Command Line Environment`. It allows you to register commands and handlers for special events and automatically sets correct handlers for command history and completion when reading a command:

```golang
// instantiate command line environment
cle := console.NewCommandLineEnvironment()
// register default exit command
cle.RegisterCommand(console.NewExitCommand("exit"))
// register a new command that takes no parameters
cle.RegisterCommand(console.NewParameterlessCommand("hello",
    func(args []string) error {
        // handler method for command execution
        console.Println("world")
    }))

// Run() enters an infinite loop for command input and execution
// if a command returns console.ErrExit() this loop will stop gracefully
if err := cle.Run(); err != nil {
    console.Fatalln(err)
}
```

In most cases you probably want to allow completion of command arguments. This can be done by passing a completion handler when instantiating a custom command. For simple command arguments you can use the default handlers of this package:

```golang
cle.RegisterCommand(console.NewCustomCommand("hello",
    // default completion handler for a fixed set of arguments (fixed order and no flags)
    console.NewFixedArgCompletion(
        // the first argument will offer completions for a fixed set of options
        console.NewOneOfArgCompletion("world", "ma'am", "sir"),
        // you can pass an arbitrary number of completion handlers here for further arguments
    ),
    func(args []string) error {
        // the completion does not enforce presence of arguments
        if len(args) > 0 {
            console.Printlnf("hello %s", args[0])
        }
    }))
```

To select a file or directory from the local file system, there is a default completion handler available. Use the withFiles flag to control whether files should be offered in the completion list:

```golang
cle.RegisterCommand(console.NewCustomCommand("cat",
    // allow user to browse the local file system for completion and include files
    console.NewFixedArgCompletion(console.NewLocalFileSystemArgCompletion(true)),
    // actual command execution handler:
    func(args []string) error {
        // always remember to check arguments
        if len(args) == 0 {
            console.Println("missing arg")
        } else {
            // show content of file...
        }
    }))
```

See `examples/command-line-env`, `examples/error-handling` and `examples/browser` for example applications.

### Customizations

See the following list for possible customizations of the `Command Line Environment`:

| Field Name | Description | Default |
| ---------- | ----------- | ------- |
| Prompt | Callback function to specify the current command prompt. Will be called every time the prompt is displayed. Use `cle.SetStaticPrompt` to set a static prompt. | `cle> ` |
| PrintOptions | Callback function to print options on double-tab. | `DefaultOptionsPrinter()` |
| ExecUnknownCommad | A handler that is called when an unknown command is executed. If set to `nil`, the execution loop will end returning an unkown command error. | Print message and continue |
| CompleteUnknownCommand | Completion handler for unknown commands. | `nil` |
| ErrorHandler | Error handler to handle errors and panics returned from commands. Will end the execution loop and pass through the error if something else than `nil` is returned. | Print error message and continue |
| RecoverPanickedCommands | If set to `true`, panics from commands are recovered and passed to `ErrorHandler`. Use `console.IsErrCommandPanicked` to recognize panics. | `true` |
| UseCommandNameCompletion | If set to `false`, no completion is available for command names. | `true` |

### Custom Completion Handlers

Completion handlers are called every time the user presses the tab key. They receive the full, parsed command as input, aswell as the index of the currently edited entry. When using completion handlers for registered commands of a command line environment, you can ignore the first entry as it will always contain the name of the corresponding command. The handler can return the full list of available options because prefix filtering for the current user input will be done automatically:

```golang
func completionHandler(cmd []string, index int) []console.CompletionOption {
    return []console.CompletionOption{
        // labelled options will be displayed with custom label in listings
        console.NewLabelledCompletionOption("Greeting", "hello", false),
        // in most cases only the ReplaceString property is required
        console.NewCompletionOption("hedgehog", false),
        console.NewCompletionOption("world", false),
    }
}
```

The `replacement` parameter of `NewLabelledCompletionOption` and `NewCompletionOption` is the actual value to be used for completion. If the user has already typed `h` and presses tab, the prefix will match `hello` and `hedgehog`, so the longest common prefix `he` is taken as completion. The user now types `l` to further specify the desired value and again presses tab. Only one candidate now matches the prefix and so the full replacement string `hello` is taken as completion. Furthermore, a whitespace character is emitted to begin the next argument because the `isPartial` parameter is set to `false`.

The first `label` parameter of `NewLabelledCompletionOption` can be set to an arbitrary value and does not affect the actual completion in any way. It is displayed instead of the actual replacement property in the options list on double-tab. This can be useful when completion is used on hierachical structures like file systems where you only want to display the file names and not the full path.

### Custom Completion Arg

You can also extend the `NewFixedArgCompletion` with custom types that implement `ArgCompletion`:

```golang
type ArgCompletion interface {
    GetCompletionOptions(currentCommand []string, entryIndex int) (options []CompletionOption)
}
```

## Applications

See [s3client](https://github.com/sbreitf1/s3client) for a real-world application using go-console.