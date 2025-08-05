# WTF Custom Command Names

WTF makes it super easy to use any command name you want!

## 🚀 **Super Simple Setup**

```bash
# One command setup - WTF handles everything!
wtf setup hey        # Creates 'hey' command  
wtf setup miko       # Creates 'miko' command
wtf setup cmd        # Creates 'cmd' command

# Then use it:
hey "compress files"
miko "git commands"
cmd "find large files"

# With platform filtering:
hey "list files" --platform linux
miko "system tools" --platform windows,macos
cmd "process management" --all-platforms
```

## 🪟 **Windows Users**

For Windows Command Prompt (the easiest way):

```cmd
# One-time setup:
wtf setup hey

# Creates hey.bat in current directory
# Then use: hey "your query"
```

**Alternative for current session:**
```cmd
doskey hey=wtf.exe $*
doskey miko=wtf.exe $*

# Now you can use: hey "compress files"
```

## 🐧 **Linux/Mac Users**

```bash
# Automatic setup:
wtf setup hey

# Or manual setup:
alias hey='wtf'
alias miko='wtf'

# Add to ~/.bashrc or ~/.zshrc to make permanent
```

## 📋 **Manage Your Aliases**

```bash
wtf alias add hey           # Add new alias
wtf alias list              # See all aliases  
wtf alias remove hey        # Remove alias
```

## 💡 **Popular Command Names**

- `hey` - Conversational ("hey, find files")
- `miko` - Personal assistant style  
- `cmd` - Short and clear
- `wtf` - The classic (What's The Function!)
- `help` - Descriptive
- `find-cmd` - Self-explanatory

## ✨ **That's It!**

No complex shell configurations, no PATH editing, no PowerShell profiles. Just run `wtf setup [name]` and you're done! 🎯
