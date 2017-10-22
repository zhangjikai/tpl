## Introduction

Tpl is a simple tool for easy managing of our file or project templates. We first add the template file or project to our template library use `save` command. And then we can fetch the saved template use `get` command in any directory.

## Installation
You can install tpl with the following command:
```
go get github.com/zhangjikai/tpl
```

## Usage

* `save,s [key] [template path]` - Save a template associated with the specified key to the library.
* `get,g [key]` - Get a template associated with the specified key from the library.
* `delete,d [key]` - Delete a template associated with the specified key from the library.
* `ls,l [prefix]` - List the keys that begin with prefix. If the prefix is not specified, it will list all keys.
* `config,c [type] [value]` - Set configurations of tpl. If no parameters are passed, the current configurations will be printed. Valid configurations:
    - `StorePath`: Specifies the directroy that stores template files.
* `push` - Call git push command based on the template library directory.
    - In order to make `push` and `pull` work properly, you need to specify a git repository as the storage directory of templates. For example:
      ```
      cd <dir>
      git clone git@xxx.xxx/templates.git
      tpl config StorePath <dir>/tempaltes
      ```
* `pull` - Call git pull command based on the template library directory.
