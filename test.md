D:\WTF2>build\wtf.exe "rsync backup files"
2026/01/17 10:36:13 Loaded semantic search: 100000 words, 3264 command embeddings
2026/01/17 10:36:13 Loaded semantic search: 100000 words, 3264 command embeddings
2026/01/17 10:36:13 Loaded semantic search: 100000 words, 3264 command embeddings
Searching for: rsync backup files

Found 5 matching command(s):

1. lz
   Description: List all files inside a '.tar.gz' compressed archive.
   Category: system

2. removepkg -copy foo-0.2.8-x86_64-1  # -> /var/log/setup/tmp/preserved_packages/foo...
   Description: remove a Slackware package, retaining a backup (uninstalled) copy

3. po4a
   Description: Update both PO files and translated documents.
   Category: system

4. rsync -avc <src> <dest>
   Description: copy files using checksum (-c) rather than time to detect if the file has changed

5. tar czvf /tmp/backup.tar.gz $(installpkg --warn foo-1.0.4-noarch-1.tgz)
   Description: create backup of files that will be overwritten when installing


D:\WTF2>build\wtf.exe "change file permissions"
2026/01/17 10:36:26 Loaded semantic search: 100000 words, 3264 command embeddings
2026/01/17 10:36:26 Loaded semantic search: 100000 words, 3264 command embeddings
2026/01/17 10:36:26 Loaded semantic search: 100000 words, 3264 command embeddings
Searching for: change file permissions

Found 5 matching command(s):

1. chmod u+x
   Description: Change the access permissions of a file or directory.
   Category: general

2. find . -type f -exec chmod 644 {} \;
   Description: find all files in the current directory and modify their permissions
   Category: filesystem

3. chcon
   Description: Change SELinux security context of a file or files/directories. See also: `secon`, `restorecon`, `semanage-fcontext`.
   Category: system

4. ssh -i <pemfile> <user>@<host>
   Description: ssh via pem file (which normally needs 0600 permissions)
   Category: networking

5. ftype
   Description: Display or modify file types used for file extension association.
   Category: system


D:\WTF2>build\wtf.exe "how to see file contents without opening"
2026/01/17 10:36:37 Loaded semantic search: 100000 words, 3264 command embeddings
2026/01/17 10:36:37 Loaded semantic search: 100000 words, 3264 command embeddings
2026/01/17 10:36:37 Loaded semantic search: 100000 words, 3264 command embeddings
Searching for: how to see file contents without opening

Found 5 matching command(s):

1. truncate -s 0 <file>
   Description: clear the contents from <file>

2. head <file>
   Description: show the first 10 lines of <file>

3. head -n <number> <file>
   Description: show the first <number> lines of <file>

4. head -c <number> <file>
   Description: show the first <number> bytes of <file>

5. tail <file>
   Description: show the last 10 lines of <file>

