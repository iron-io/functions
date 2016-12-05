# Using dotnet with functions

Make sure you downloaded and installed [dotnet](https://www.microsoft.com/net/core). Now create an empty dotnet project in the directory of your function:

```bash
dotnet new
```

By default dotnet creates a ```Program.cs``` file with a main method. To make it work with IronFunction's `fn` tool please rename it to ```func.cs```.
Now change the code as you desire to do whatever magic you need it to do. Once done you can now create an iron function out of it.

## Creating an IronFunction
Simply run

```bash
fn init <username>/<funcname>
```

This will create the ```func.yaml``` file required by functions, which can be built by running:

## Deploying

```bash
fn deploy <app_name>
```

## Testing

```bash
fn run
```

## Calling

```bash
fn call <app_name> <funcname>
```