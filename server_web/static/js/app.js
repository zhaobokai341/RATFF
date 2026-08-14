        const { createApp, ref, computed } = Vue;

        // Extract path prefix from current URL
        const pathParts = window.location.pathname.split('/').filter(Boolean);
        const basePath = pathParts.length > 0 ? '/' + pathParts[0] : '';
        createApp({
            delimiters: ['${', '}'],
            components: {
                'lang-switcher': LangSwitcher
            },
            setup() {
                var lang = '{{.lang}}' || 'zh';
                var labels = messages[lang] || messages.zh;

                const clients = ref([]);
                const selectedClient = ref('');
                const command = ref('');
                const cdDir = ref('');
                const currentDir = ref('');
                const bgOutputFile = ref('');
                const output = ref(labels.exec_waiting);
                const outputType = ref('loading');
                const outputStdout = ref('');
                const outputStderr = ref('');
                const outputExitCode = ref(null);
                const executing = ref(false);
                const deleteTarget = ref(null);
                const toast = ref('');
                const toastType = ref('');
                const sysinfoTarget = ref(null);
                const sysinfoData = ref({});
                const sysinfoLoading = ref(false);
                const selectedFields = ref([]);
                const availableFields = ['host', 'cpu', 'memory', 'swap_memory', 'partition', 'io_disk', 'interfaces', 'io_network', 'processes'];

                const showFileManager = ref(false);
                const fmPath = ref('');
                const fmFiles = ref([]);
                const fmLoading = ref(false);
                const fmDeleteTarget = ref(null);
                const fmMoveTarget = ref(false);
                const fmMoveOrigin = ref('');
                const fmMoveDest = ref('');
                const fmPropertiesTarget = ref(null);

                const statusText = computed(() => clients.value.length > 0 ? labels.status_connected : labels.status_disconnected);

                const api = async (url, opts = {}) => {
                    const fullUrl = basePath + url;
                    const res = await fetch(fullUrl, { headers: { 'Content-Type': 'application/json' }, ...opts });
                    return res.json();
                };

                const loadClients = async () => {
                    try {
                        const data = await api('/api/clients');
                        clients.value = data.clients || [];
                    } catch (e) {
                        showToast(labels.toast_failed_load_clients, 'error');
                    }
                };

                const sendCommand = async (clientId, cmd, payload) => {
                    executing.value = true;
                    output.value = labels.exec_executing;
                    outputType.value = 'loading';
                    outputStdout.value = '';
                    outputStderr.value = '';
                    outputExitCode.value = null;
                    try {
                        const data = await api('/api/command', {
                            method: 'POST',
                            body: JSON.stringify({ client_id: clientId, command: cmd, payload: payload || {} })
                        });
                        if (data.error) {
                            output.value = 'Error: ' + data.error;
                            outputType.value = 'error';
                        } else if (data.response && data.response.payload) {
                            var p = data.response.payload;
                            if (p.stdout !== undefined || p.stderr !== undefined || p.exit_code !== undefined) {
                                outputStdout.value = p.stdout || '';
                                outputStderr.value = p.stderr || '';
                                outputExitCode.value = p.exit_code !== undefined ? p.exit_code : null;
                                outputType.value = (p.exit_code === 0) ? 'success' : 'error';
                            } else {
                                output.value = p.output || (p.error ? '[Error] ' + p.error : '[No output]');
                                outputType.value = p.error ? 'error' : 'success';
                            }
                        } else {
                            output.value = JSON.stringify(data, null, 2);
                            outputType.value = 'success';
                        }
                    } catch (e) {
                        output.value = 'Request failed: ' + e.message;
                        outputType.value = 'error';
                    } finally {
                        executing.value = false;
                    }
                };

                const execCommand = function() {
                    if (!selectedClient.value) return showToast(labels.toast_select_client, 'error');
                    if (!command.value.trim()) return showToast(labels.toast_enter_command, 'error');
                    sendCommand(selectedClient.value, 'shell_exec', { cmd: command.value.trim() });
                };

                const execBgCommand = function() {
                    if (!selectedClient.value) return showToast(labels.toast_select_client, 'error');
                    if (!command.value.trim()) return showToast(labels.toast_enter_command, 'error');
                    executing.value = true;
                    output.value = labels.exec_executing;
                    outputType.value = 'loading';
                    var payload = { cmd: command.value.trim() };
                    if (bgOutputFile.value.trim()) {
                        payload.output_file = bgOutputFile.value.trim();
                    }
                    try {
                        api('/api/command', {
                            method: 'POST',
                            body: JSON.stringify({ client_id: selectedClient.value, command: 'shell_exec_bg', payload: payload })
                        }).then(function(data) {
                            if (data.response && data.response.payload && data.response.payload.status === 'started') {
                                var outFile = data.response.payload.output_file;
                                var msg = labels.toast_bg_started + ': ' + command.value.trim();
                                if (outFile) {
                                    msg += ' (output: ' + outFile + ')';
                                }
                                output.value = msg;
                                outputType.value = 'success';
                                showToast(labels.toast_bg_started, 'success');
                            } else if (data.error) {
                                output.value = 'Error: ' + data.error;
                                outputType.value = 'error';
                            }
                        }).catch(function(e) {
                            output.value = 'Request failed: ' + e.message;
                            outputType.value = 'error';
                        });
                    } finally {
                        executing.value = false;
                    }
                };

                const execCd = function() {
                    if (!selectedClient.value) return showToast(labels.toast_select_client, 'error');
                    if (!cdDir.value.trim()) return showToast(labels.toast_enter_dir, 'error');
                    executing.value = true;
                    try {
                        api('/api/command', {
                            method: 'POST',
                            body: JSON.stringify({ client_id: selectedClient.value, command: 'cd', payload: { dir: cdDir.value.trim() } })
                        }).then(function(data) {
                            if (data.response && data.response.payload) {
                                var p = data.response.payload;
                                if (p.current_dir) {
                                    currentDir.value = p.current_dir;
                                    output.value = labels.toast_cd_success + ': ' + p.current_dir;
                                    outputType.value = 'success';
                                    showToast(labels.toast_cd_success, 'success');
                                } else if (p.error) {
                                    output.value = labels.toast_cd_failed + ': ' + p.error;
                                    outputType.value = 'error';
                                    showToast(labels.toast_cd_failed, 'error');
                                }
                            } else if (data.error) {
                                output.value = labels.toast_cd_failed + ': ' + data.error;
                                outputType.value = 'error';
                                showToast(labels.toast_cd_failed, 'error');
                            }
                        }).catch(function(e) {
                            output.value = 'Request failed: ' + e.message;
                            outputType.value = 'error';
                        });
                    } finally {
                        executing.value = false;
                    }
                };

                const selectAndInfo = function(id) {
                    selectedClient.value = id;
                    sendCommand(id, 'system_info', {});
                };

                const showDelete = function(id) { deleteTarget.value = id; };

                const showSysInfo = function(id) {
                    sysinfoTarget.value = id;
                    sysinfoData.value = {};
                    selectedFields.value = [];
                };

                const fetchSysInfo = async function() {
                    if (!sysinfoTarget.value) return;
                    sysinfoLoading.value = true;
                    sysinfoData.value = {};
                    try {
                        var fields = selectedFields.value.length > 0 ? selectedFields.value : availableFields;
                        var data = await api('/api/command', {
                            method: 'POST',
                            body: JSON.stringify({
                                client_id: sysinfoTarget.value,
                                command: 'system_info',
                                payload: { fields: fields }
                            })
                        });
                        if (data.response && data.response.payload) {
                            sysinfoData.value = data.response.payload;
                        } else if (data.error) {
                            showToast('Error: ' + data.error, 'error');
                        }
                    } catch (e) {
                        showToast('Request failed: ' + e.message, 'error');
                    } finally {
                        sysinfoLoading.value = false;
                    }
                };

                const formatBytes = function(bytes) {
                    if (!bytes || bytes === 0) return '0 B';
                    var units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
                    var unitIndex = 0;
                    var value = parseFloat(bytes);
                    while (value >= 1024 && unitIndex < units.length - 1) {
                        value /= 1024;
                        unitIndex++;
                    }
                    return value.toFixed(1) + ' ' + units[unitIndex];
                };

                const formatUptime = function(seconds) {
                    if (!seconds || seconds === 0) return '0s';
                    var days = Math.floor(seconds / 86400);
                    var hours = Math.floor((seconds % 86400) / 3600);
                    var minutes = Math.floor((seconds % 3600) / 60);
                    var secs = seconds % 60;
                    var result = '';
                    if (days > 0) result += days + 'd ';
                    if (hours > 0) result += hours + 'h ';
                    if (minutes > 0) result += minutes + 'm ';
                    result += secs + 's';
                    return result;
                };

                const formatFieldName = function(key) {
                    var nameMap = {
                        'total': 'Total',
                        'used': 'Used',
                        'available': 'Available',
                        'free': 'Free',
                        'active': 'Active',
                        'inactive': 'Inactive',
                        'buffers': 'Buffers',
                        'cached': 'Cached',
                        'sin': 'Sin',
                        'sout': 'Sout',
                        'used_percent': 'Used Percent',
                        'hostname': 'Hostname',
                        'os': 'OS',
                        'platform': 'Platform',
                        'platform_version': 'Platform Version',
                        'platform_family': 'Platform Family',
                        'kernel_version': 'Kernel Version',
                        'kernel_arch': 'Kernel Arch',
                        'uptime': 'Uptime',
                        'boot_time': 'Boot Time',
                        'procs': 'Processes',
                        'model_name': 'Model',
                        'cores': 'Cores',
                        'mhz': 'MHz',
                        'cache_size': 'Cache Size',
                        'vendor_id': 'Vendor',
                        'read_count': 'Read Count',
                        'write_count': 'Write Count',
                        'read_bytes': 'Read Bytes',
                        'write_bytes': 'Write Bytes',
                        'read_time': 'Read Time',
                        'write_time': 'Write Time',
                        'bytes_sent': 'Bytes Sent',
                        'bytes_recv': 'Bytes Recv',
                        'packets_sent': 'Packets Sent',
                        'packets_recv': 'Packets Recv',
                        'errin': 'Errors In',
                        'errout': 'Errors Out',
                        'dropin': 'Dropped In',
                        'dropout': 'Dropped Out',
                        'mtu': 'MTU',
                        'flags': 'Flags',
                        'addresses': 'Addresses'
                    };
                    if (nameMap[key]) {
                        return nameMap[key];
                    }
                    return key.split('_').map(function(word) {
                        return word.charAt(0).toUpperCase() + word.slice(1);
                    }).join(' ');
                };

                const formatFieldValue = function(key, value) {
                    if (value === undefined || value === null) return 'N/A';
                    
                    switch (key) {
                        case 'used_percent':
                            return typeof value === 'number' ? value.toFixed(1) + '%' : 'N/A';
                        case 'uptime':
                            return formatUptime(value);
                        case 'boot_time':
                            return value ? new Date(value * 1000).toLocaleString() : 'N/A';
                        case 'cache_size':
                            return typeof value === 'number' ? value + ' KB' : 'N/A';
                        case 'mhz':
                            return typeof value === 'number' ? value.toFixed(2) : 'N/A';
                        case 'read_time':
                        case 'write_time':
                            return typeof value === 'number' ? value + ' ms' : 'N/A';
                        case 'flags':
                            return Array.isArray(value) ? value.join(', ') : (value || 'N/A');
                        case 'addresses':
                            return Array.isArray(value) ? value.join(', ') : (value || 'N/A');
                        default:
                            if (typeof value === 'number' && (key.includes('bytes') || key === 'total' || key === 'used' || key === 'free' || key === 'available' || key === 'active' || key === 'inactive' || key === 'buffers' || key === 'cached' || key === 'sin' || key === 'sout' || key.includes('sent') || key.includes('recv'))) {
                                return formatBytes(value);
                            }
                            return value || 'N/A';
                    }
                };

                const confirmDelete = async function() {
                    var id = deleteTarget.value;
                    deleteTarget.value = null;
                    output.value = labels.toast_exit_sent_prefix + ' ' + id + '...';
                    outputType.value = 'loading';
                    try {
                        var data = await api('/api/command', {
                            method: 'POST',
                            body: JSON.stringify({ client_id: id, command: 'exit' })
                        });
                        if (data.error) {
                            output.value = 'Error: ' + data.error;
                            outputType.value = 'error';
                        } else {
                            output.value = '"' + id + '" ' + labels.toast_exit_sent_prefix + '.';
                            outputType.value = 'success';
                            showToast(labels.toast_delete_success, 'success');
                            setTimeout(loadClients, 1000);
                        }
                    } catch (e) {
                        output.value = 'Request failed: ' + e.message;
                        outputType.value = 'error';
                    }
                };

                const showToast = function(msg, type) {
                    toast.value = msg;
                    toastType.value = type === 'success' ? 'bg-green-500 text-white' : 'bg-red-500 text-white';
                    setTimeout(function() { toast.value = ''; }, 3000);
                };

                const logout = function() {
                    window.location.href = '/logout';
                };

                const showFileManagerFor = function(id) {
                    selectedClient.value = id;
                    showFileManager.value = true;
                    fmPath.value = '';
                    loadFiles();
                };

                const loadFiles = async function() {
                    if (!selectedClient.value) return showToast(labels.fm_toast_select_client, 'error');
                    fmLoading.value = true;
                    try {
                        var data = await api('/api/file/list', {
                            method: 'POST',
                            body: JSON.stringify({ client_id: selectedClient.value, path: fmPath.value })
                        });
                        if (data.response) {
                            var resp = data.response;
                            fmPath.value = resp.path || fmPath.value;
                            fmFiles.value = resp.files || [];
                        } else if (data.error) {
                            showToast(labels.fm_load_failed.replace('{error}', data.error), 'error');
                            fmFiles.value = [];
                        }
                    } catch (e) {
                        showToast(labels.fm_load_failed.replace('{error}', e.message), 'error');
                        fmFiles.value = [];
                    } finally {
                        fmLoading.value = false;
                    }
                };

                const fmGoBack = function() {
                    if (!fmPath.value || fmPath.value === '/') {
                        fmPath.value = '';
                    } else {
                        var parts = fmPath.value.split(/[\/\\]/).filter(function(p) { return p; });
                        parts.pop();
                        fmPath.value = parts.length > 0 ? (fmPath.value.includes('\\') ? parts.join('\\') : '/' + parts.join('/')) : '';
                    }
                    loadFiles();
                };

                const handleFileClick = function(f) {
                    if (f.is_dir) {
                        var sep = fmPath.value.includes('\\') ? '\\' : '/';
                        var newPath = fmPath.value ? (fmPath.value.endsWith(sep) ? fmPath.value + f.name : fmPath.value + sep + f.name) : f.name;
                        fmPath.value = newPath;
                        loadFiles();
                    }
                };

                const getDisplayName = function(f) {
                    var name = f.name;
                    if (f.hidden) {
                        name = labels.fm_hidden_prefix + name;
                    }
                    if (f.link_target) {
                        name += ' -> ' + f.link_target;
                    }
                    return name;
                };

                const getFileTypeIcon = function(f) {
                    if (f.type === 'symlink' || f.type === 'shortcut') return '🔗';
                    if (f.is_dir) return '📁';
                    return '📄';
                };

                const getFileTypeLabel = function(f) {
                    if (f.type === 'symlink') return labels.fm_type_symlink;
                    if (f.type === 'shortcut') return labels.fm_type_shortcut;
                    if (f.is_dir) return labels.fm_type_directory;
                    return labels.fm_type_file;
                };

                const formatFileSize = function(bytes) {
                    if (!bytes || bytes === 0) return '0 B';
                    var units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
                    var unitIndex = 0;
                    var value = parseFloat(bytes);
                    while (value >= 1024 && unitIndex < units.length - 1) {
                        value /= 1024;
                        unitIndex++;
                    }
                    return value.toFixed(1) + ' ' + units[unitIndex];
                };

                const formatModTime = function(unixTime) {
                    if (!unixTime) return 'N/A';
                    var d = new Date(unixTime * 1000);
                    return d.toLocaleString();
                };

                const showDeleteFile = function(f) {
                    var sep = fmPath.value.includes('\\') ? '\\' : '/';
                    var fullPath = fmPath.value ? (fmPath.value.endsWith(sep) ? fmPath.value + f.name : fmPath.value + sep + f.name) : f.name;
                    fmDeleteTarget.value = fullPath;
                };

                const confirmDeleteFile = async function() {
                    var path = fmDeleteTarget.value;
                    fmDeleteTarget.value = null;
                    try {
                        var data = await api('/api/file/delete', {
                            method: 'POST',
                            body: JSON.stringify({ client_id: selectedClient.value, path: path })
                        });
                        if (data.error) {
                            showToast(labels.fm_delete_failed.replace('{error}', data.error), 'error');
                        } else {
                            showToast(labels.fm_delete_success, 'success');
                            loadFiles();
                        }
                    } catch (e) {
                        showToast(labels.fm_delete_failed.replace('{error}', e.message), 'error');
                    }
                };

                const showMoveFile = function(f) {
                    var sep = fmPath.value.includes('\\') ? '\\' : '/';
                    var fullPath = fmPath.value ? (fmPath.value.endsWith(sep) ? fmPath.value + f.name : fmPath.value + sep + f.name) : f.name;
                    fmMoveOrigin.value = fullPath;
                    fmMoveDest.value = '';
                    fmMoveTarget.value = true;
                };

                const resolvePath = function(base, target) {
                    if (!target) return base;

                    var isWindows = target.includes(':') || (base && base.includes(':'));
                    var sep = isWindows ? '\\' : '/';

                    if (isWindows && target.includes(':') && target[1] === ':') {
                        var resolved = normalizePathParts(target.split(/[\/\\]/), sep);
                        return resolved.join(sep);
                    }

                    if (target.startsWith('/') || (isWindows && target.startsWith('\\'))) {
                        var resolved = normalizePathParts(target.split(/[\/\\]/), sep);
                        return isWindows ? resolved.join(sep) : sep + resolved.join(sep);
                    }

                    var basePath = base || '';
                    var baseParts = basePath.split(/[\/\\]/).filter(function(p) { return p; });
                    var targetParts = target.split(/[\/\\]/).filter(function(p) { return p; });
                    var allParts = baseParts.concat(targetParts);
                    var resolved = normalizePathParts(allParts, sep);

                    if (!isWindows && basePath.startsWith('/')) {
                        return sep + resolved.join(sep);
                    }
                    if (isWindows && basePath.includes(':')) {
                        var drive = basePath.substring(0, 2);
                        return drive + sep + resolved.join(sep);
                    }
                    return resolved.join(sep);
                };

                const normalizePathParts = function(parts, sep) {
                    var stack = [];
                    for (var i = 0; i < parts.length; i++) {
                        var part = parts[i];
                        if (part === '.' || part === '') {
                            continue;
                        } else if (part === '..') {
                            if (stack.length > 0) {
                                stack.pop();
                            }
                        } else {
                            stack.push(part);
                        }
                    }
                    return stack;
                };

                const confirmMoveFile = async function() {
                    if (!fmMoveDest.value.trim()) return;
                    var origin = fmMoveOrigin.value;
                    var dest = fmMoveDest.value.trim();
                    fmMoveTarget.value = false;

                    dest = resolvePath(fmPath.value, dest);

                    try {
                        var data = await api('/api/file/move', {
                            method: 'POST',
                            body: JSON.stringify({ client_id: selectedClient.value, origin_path: origin, new_path: dest })
                        });
                        if (data.error) {
                            showToast(labels.fm_move_failed.replace('{error}', data.error), 'error');
                        } else {
                            showToast(labels.fm_move_success, 'success');
                            loadFiles();
                        }
                    } catch (e) {
                        showToast(labels.fm_move_failed.replace('{error}', e.message), 'error');
                    }
                };

                const showProperties = function(f) {
                    fmPropertiesTarget.value = f;
                };

                loadClients();

                return {
                    clients: clients,
                    selectedClient: selectedClient,
                    command: command,
                    cdDir: cdDir,
                    currentDir: currentDir,
                    bgOutputFile: bgOutputFile,
                    output: output,
                    outputType: outputType,
                    outputStdout: outputStdout,
                    outputStderr: outputStderr,
                    outputExitCode: outputExitCode,
                    executing: executing,
                    deleteTarget: deleteTarget,
                    toast: toast,
                    toastType: toastType,
                    statusText: statusText,
                    labels: labels,
                    loadClients: loadClients,
                    execCommand: execCommand,
                    execBgCommand: execBgCommand,
                    execCd: execCd,
                    selectAndInfo: selectAndInfo,
                    showDelete: showDelete,
                    confirmDelete: confirmDelete,
                    logout: logout,
                    sysinfoTarget: sysinfoTarget,
                    sysinfoData: sysinfoData,
                    sysinfoLoading: sysinfoLoading,
                    selectedFields: selectedFields,
                    availableFields: availableFields,
                    showSysInfo: showSysInfo,
                    fetchSysInfo: fetchSysInfo,
                    formatBytes: formatBytes,
                    formatUptime: formatUptime,
                    formatFieldName: formatFieldName,
                    formatFieldValue: formatFieldValue,
                    showFileManager: showFileManager,
                    fmPath: fmPath,
                    fmFiles: fmFiles,
                    fmLoading: fmLoading,
                    fmDeleteTarget: fmDeleteTarget,
                    fmMoveTarget: fmMoveTarget,
                    fmMoveOrigin: fmMoveOrigin,
                    fmMoveDest: fmMoveDest,
                    fmPropertiesTarget: fmPropertiesTarget,
                    showFileManagerFor: showFileManagerFor,
                    loadFiles: loadFiles,
                    fmGoBack: fmGoBack,
                    handleFileClick: handleFileClick,
                    getDisplayName: getDisplayName,
                    getFileTypeIcon: getFileTypeIcon,
                    getFileTypeLabel: getFileTypeLabel,
                    formatFileSize: formatFileSize,
                    formatModTime: formatModTime,
                    showDeleteFile: showDeleteFile,
                    confirmDeleteFile: confirmDeleteFile,
                    showMoveFile: showMoveFile,
                    confirmMoveFile: confirmMoveFile,
                    showProperties: showProperties
                };
            }
        }).mount('#app');