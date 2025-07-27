# Lua refactoring - pattern matching for git remotes

**Date:** 2025-07-26 00:57:00
**Tags:** lua,refactoring,git,pattern-matching
**Session ID:** 2025-07-26_00-57-00_Lua_refactoring_-_pattern_matching_for_git_remotes

## Prompt

```
longer text test prompt let's focus on remap.lua instead of using this if statements I want to look for a word in remote_url for github or gitlab 

   if remote_url:match("github.com") then
      repo_path = remote_url:match("git@github.com:(.+)%.git") or remote_url:match("https://github.com/(.+)%.git")
      base_url = "https://github.com"
   elseif remote_url:match("gitlab.com") then
      repo_path = remote_url:match("git@gitlab.com:(.+)%.git") or remote_url:match("https://gitlab.com/(.+)%.git")
      base_url = "https://gitlab.com"
   else
      print("Error: Unsupported remote host\!")
      return
   end
```

## Context

- Working Directory: /home/dariuszw/.Dotfiles
- Git Branch: Linux
- Git Status: 1 modified files

## Notes

<!-- Add implementation notes, results, or follow-up actions here -->

