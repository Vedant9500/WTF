package nlp

// getCommandHints returns actual command names based on detected actions and targets
// This bridges the gap between natural language and command names
func (pq *ProcessedQuery) getCommandHints() []string {
	var hints []string

	// Helper functions for checking conditions
	hasAction := func(action string) bool {
		for _, a := range pq.Actions {
			if a == action {
				return true
			}
		}
		for _, k := range pq.Keywords {
			if k == action {
				return true
			}
		}
		return false
	}

	hasTarget := func(targets ...string) bool {
		for _, t := range pq.Targets {
			for _, target := range targets {
				if t == target {
					return true
				}
			}
		}
		for _, k := range pq.Keywords {
			for _, target := range targets {
				if k == target {
					return true
				}
			}
		}
		return false
	}

	hasKeyword := func(keywords ...string) bool {
		for _, k := range pq.Keywords {
			for _, keyword := range keywords {
				if k == keyword {
					return true
				}
			}
		}
		return false
	}

	// Aggregate hints from various categories
	hints = append(hints, pq.getDirectoryHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getFileHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getCompressionHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getArchiveHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getDownloadHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getProcessHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getNetworkHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getDiskHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getTextHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getEditorHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getLogHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getFindReplaceHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getSearchHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getInstallHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getPermissionHints(hasAction, hasTarget, hasKeyword)...)
	hints = append(hints, pq.getRemoteHints(hasAction, hasTarget, hasKeyword)...)

	return hints
}

func (pq *ProcessedQuery) getDirectoryHints(hasAction func(string) bool, hasTarget, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasTarget("directory", "folder", "directories", "folders", "dir") ||
		hasKeyword("directory", "folder", "directories", "folders", "dir") {
		if hasAction("create") || hasAction("make") || hasAction("new") ||
			hasKeyword("create", "make", "new") || pq.Intent == IntentCreate {
			hints = append(hints, "mkdir")
		}
		if hasAction("delete") || hasAction("remove") || pq.Intent == IntentDelete {
			hints = append(hints, "rmdir", "rm")
		}
		if hasAction("list") || hasAction("show") || pq.Intent == IntentFind {
			hints = append(hints, "ls", "dir")
		}
	}
	return hints
}

func (pq *ProcessedQuery) getFileHints(hasAction func(string) bool, hasTarget, hasKeyword func(...string) bool) []string {
	if !hasTarget("file", "files") && !hasKeyword("file", "files") {
		return nil
	}
	return pq.collectFileActionHints(hasAction, hasKeyword)
}

// collectFileActionHints returns hints based on file-related actions
func (pq *ProcessedQuery) collectFileActionHints(hasAction func(string) bool, hasKeyword func(...string) bool) []string {
	var hints []string
	hints = append(hints, pq.getFileCopyMoveHints(hasAction)...)
	hints = append(hints, pq.getFileSearchViewHints(hasAction)...)
	hints = append(hints, pq.getFileEditCompressHints(hasAction, hasKeyword)...)
	return hints
}

// getFileCopyMoveHints returns hints for copy, move, delete operations
func (pq *ProcessedQuery) getFileCopyMoveHints(hasAction func(string) bool) []string {
	var hints []string
	if hasAction("copy") {
		hints = append(hints, "cp")
	}
	if hasAction("move") || hasAction("rename") {
		hints = append(hints, "mv")
	}
	if hasAction("delete") || hasAction("remove") {
		hints = append(hints, "rm")
	}
	return hints
}

// getFileSearchViewHints returns hints for find, search, view operations
func (pq *ProcessedQuery) getFileSearchViewHints(hasAction func(string) bool) []string {
	var hints []string
	if hasAction("find") || hasAction("search") || hasAction("locate") {
		hints = append(hints, "find", "grep")
	}
	if hasAction("view") || hasAction("show") || hasAction("read") || hasAction("see") {
		hints = append(hints, "cat", "less", "more")
	}
	return hints
}

// getFileEditCompressHints returns hints for edit, compress, extract, download operations
func (pq *ProcessedQuery) getFileEditCompressHints(hasAction func(string) bool, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasAction("edit") || hasKeyword("edit") {
		hints = append(hints, "vim", "nano", "vi")
	}
	if hasAction("compress") || hasAction("archive") || hasAction("zip") {
		hints = append(hints, "tar", "zip", "gzip")
	}
	if hasAction("extract") || hasAction("unzip") || hasAction("decompress") {
		hints = append(hints, "tar", "unzip", "gunzip")
	}
	if hasAction("download") {
		hints = append(hints, "wget", "curl")
	}
	return hints
}

func (pq *ProcessedQuery) getCompressionHints(hasAction func(string) bool, _, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasAction("compress") || hasKeyword("compress") {
		hints = append(hints, "tar", "zip", "gzip")
	}
	return hints
}

func (pq *ProcessedQuery) getArchiveHints(hasAction func(string) bool, hasTarget, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasTarget("archive", "archives") || hasKeyword("archive", "archives") {
		if hasAction("extract") || hasKeyword("extract", "unpack", "decompress") {
			hints = append(hints, "tar", "unzip", "gunzip")
		}
		if hasAction("create") || hasAction("compress") {
			hints = append(hints, "tar", "zip")
		}
		hints = append(hints, "tar")
	}
	if hasAction("extract") || hasKeyword("extract", "unzip", "decompress", "unpack") {
		hints = append(hints, "tar", "unzip", "gunzip")
	}
	return hints
}

func (pq *ProcessedQuery) getDownloadHints(hasAction func(string) bool, _, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasAction("download") || hasAction("fetch") || hasKeyword("download", "fetch") {
		hints = append(hints, "wget", "curl")
	}
	return hints
}

func (pq *ProcessedQuery) getProcessHints(hasAction func(string) bool, hasTarget, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasTarget("process", "processes", "task", "tasks") || hasKeyword("process", "processes") {
		if hasAction("list") || hasAction("show") || pq.Intent == IntentFind {
			hints = append(hints, "ps", "top", "htop")
		}
		if hasAction("kill") || hasAction("stop") || hasAction("terminate") {
			hints = append(hints, "kill", "pkill")
		}
	}
	return hints
}

func (pq *ProcessedQuery) getNetworkHints(_ func(string) bool, hasTarget, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasTarget("network", "connection", "connections", "port", "ports") ||
		hasKeyword("network", "connections", "tools") {
		hints = append(hints, "netstat", "ss", "ifconfig", "ip")
	}
	return hints
}

func (pq *ProcessedQuery) getDiskHints(_ func(string) bool, hasTarget, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasTarget("disk", "space", "storage") || hasKeyword("disk", "space", "storage", "usage") {
		hints = append(hints, "df", "du")
	}
	return hints
}

func (pq *ProcessedQuery) getTextHints(hasAction func(string) bool, hasTarget, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasTarget("text") || hasKeyword("text") {
		if hasAction("search") || hasAction("find") || hasKeyword("search") {
			hints = append(hints, "grep", "awk", "sed")
		}
		if hasAction("edit") || hasKeyword("edit", "editor") {
			hints = append(hints, "vim", "nano", "vi")
		}
		if hasKeyword("processing", "process") {
			hints = append(hints, "sed", "awk", "grep", "cut", "sort")
		}
	}
	return hints
}

func (pq *ProcessedQuery) getEditorHints(hasAction func(string) bool, hasTarget, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasKeyword("editor") || (hasAction("edit") && !hasTarget("file", "files")) {
		hints = append(hints, "vim", "nano", "vi", "emacs")
	}
	return hints
}

func (pq *ProcessedQuery) getLogHints(_ func(string) bool, hasTarget, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasTarget("log", "logs") || hasKeyword("log", "logs") {
		if hasKeyword("analysis", "analyze", "search", "find") {
			hints = append(hints, "grep", "awk", "tail", "less")
		} else {
			hints = append(hints, "tail", "less", "grep")
		}
	}
	return hints
}

func (pq *ProcessedQuery) getFindReplaceHints(hasAction func(string) bool, _, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasKeyword("replace") || (hasAction("find") && hasKeyword("replace")) {
		hints = append(hints, "sed", "awk")
	}
	return hints
}

func (pq *ProcessedQuery) getSearchHints(hasAction func(string) bool, _, hasKeyword func(...string) bool) []string {
	var hints []string
	if (hasAction("search") || hasKeyword("search")) && hasKeyword("files", "text") {
		hints = append(hints, "grep", "find")
	}
	return hints
}

func (pq *ProcessedQuery) getInstallHints(hasAction func(string) bool, hasTarget, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasAction("install") || hasKeyword("install") || pq.Intent == IntentInstall {
		if hasTarget("package", "packages") || hasKeyword("package", "packages") {
			hints = append(hints, "apt", "yum", "brew", "pip", "npm")
		} else {
			hints = append(hints, "apt", "pip", "npm")
		}
	}
	return hints
}

func (pq *ProcessedQuery) getPermissionHints(hasAction func(string) bool, hasTarget, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasTarget("permission", "permissions") || hasAction("chmod") || hasKeyword("permissions") {
		hints = append(hints, "chmod", "chown")
	}
	return hints
}

func (pq *ProcessedQuery) getRemoteHints(hasAction func(string) bool, hasTarget, hasKeyword func(...string) bool) []string {
	var hints []string
	if hasTarget("server", "remote") || hasAction("ssh") || hasKeyword("remote", "server") {
		hints = append(hints, "ssh", "scp", "rsync")
	}
	return hints
}
