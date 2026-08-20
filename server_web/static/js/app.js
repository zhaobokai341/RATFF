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
                var lang = window.APP_LANG || 'zh';
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
                const fmCopyTarget = ref(false);
                const fmCopyOrigin = ref('');
                const fmCopyDest = ref('');
                const fmPropertiesTarget = ref(null);
                const transferTask = ref(null);
                const showUploadMenu = ref(false);

                const screencapTarget = ref(null);
                const screencapFormat = ref('png');
                const screencapQuality = ref(90);
                const screencapDisplay = ref(0);
                const screencapLoading = ref(false);
                const screencapImageData = ref('');
                const screencapImageFormat = ref('png');
                const screencapWidth = ref(0);
                const screencapHeight = ref(0);
                const screencapDisplayIndex = ref(0);
                const screencapDisplayCount = ref(0);
                const screencapAutoDownload = ref(false);

                const publicipTarget = ref(null);
                const publicipData = ref({});
                const publicipLoading = ref(false);

                const updateTarget = ref(null);
                const updateSelectedFile = ref(null);
                const updateLoading = ref(false);

                const moreMenuId = ref(null);
                const moreMenuTop = ref(0);
                const moreMenuLeft = ref(0);

                const statusText = computed(() => clients.value.length > 0 ? labels.status_connected : labels.status_disconnected);

                const api = async (url, opts = {}) => {
                    const fullUrl = basePath + url;
                    const headers = {};
                    if (!opts.body || !(opts.body instanceof FormData)) {
                        headers['Content-Type'] = 'application/json';
                    }
                    const res = await fetch(fullUrl, { headers: headers, ...opts });
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
                    sendCommand(id, 'system_info', { fields: availableFields });
                };

                const showDelete = function(id) { deleteTarget.value = id; };

                const toggleMoreMenu = function(id, event) {
                    if (moreMenuId.value === id) {
                        moreMenuId.value = null;
                    } else {
                        moreMenuId.value = id;
                        var rect = event.target.getBoundingClientRect();
                        moreMenuTop.value = rect.bottom + 4;
                        moreMenuLeft.value = rect.right - 160;
                    }
                };

                const showSysInfo = function(id) {
                    sysinfoTarget.value = id;
                    sysinfoData.value = {};
                    selectedFields.value = [];
                };

                const showScreenCapture = function(id) {
                    screencapTarget.value = id;
                    screencapImageData.value = '';
                    screencapLoading.value = false;
                    screencapFormat.value = 'png';
                    screencapQuality.value = 90;
                    screencapDisplay.value = 0;
                    screencapAutoDownload.value = false;
                };

                const showPublicIP = function(id) {
                    publicipTarget.value = id;
                    publicipData.value = {};
                    publicipLoading.value = false;
                };

                const extractAPISource = function(url) {
                    if (url.indexOf('ip-api.com') !== -1) return 'ip-api.com';
                    if (url.indexOf('ipinfo.io') !== -1) return 'ipinfo.io';
                    if (url.indexOf('httpbin.org') !== -1) return 'httpbin.org';
                    return url;
                };

                const normalizeIPData = function(rawData, apiSource) {
                    var normalized = {};
                    var mapping = {};

                    if (apiSource === 'ip-api.com') {
                        mapping = {
                            ip: 'query', continent: 'continent', country: 'country',
                            country_code: 'countryCode', region: 'region', region_name: 'regionName',
                            city: 'city', district: 'district', zip: 'zip',
                            latitude: 'lat', longitude: 'lon', timezone: 'timezone',
                            isp: 'isp', org: 'org', as: 'as'
                        };
                    } else if (apiSource === 'ipinfo.io') {
                        mapping = {
                            ip: 'ip', country: 'country', country_code: 'country',
                            region: 'region', city: 'city', zip: 'postal',
                            timezone: 'timezone', isp: 'org', org: 'org'
                        };
                    } else if (apiSource === 'httpbin.org') {
                        mapping = { ip: 'origin' };
                    }

                    for (var stdKey in mapping) {
                        var rawKey = mapping[stdKey];
                        if (rawData[rawKey] !== undefined) {
                            normalized[stdKey] = rawData[rawKey];
                        }
                    }

                    if (apiSource === 'ipinfo.io' && rawData.loc) {
                        var parts = rawData.loc.split(',');
                        if (parts.length === 2) {
                            normalized.latitude = parseFloat(parts[0]);
                            normalized.longitude = parseFloat(parts[1]);
                        }
                    }

                    return normalized;
                };

                const fetchPublicIP = async function() {
                    if (!publicipTarget.value) {
                        showToast(labels.publicip_failed, 'error');
                        return;
                    }
                    publicipLoading.value = true;
                    publicipData.value = {};
                    try {
                        const data = await api('/api/public-ip', {
                            method: 'POST',
                            body: JSON.stringify({
                                client_id: publicipTarget.value
                            })
                        });
                        if (data.error) {
                            showToast(labels.publicip_failed + ': ' + data.error, 'error');
                        } else if (data.response) {
                            var payload = data.response;
                            if (payload.error) {
                                showToast(labels.publicip_failed + ': ' + payload.error, 'error');
                            } else {
                                var result = {};
                                for (var apiURL in payload) {
                                    var apiData = payload[apiURL];
                                    if (apiData.error) {
                                        result[apiURL] = { error: apiData.error };
                                    } else {
                                        var apiSource = extractAPISource(apiURL);
                                        result[apiURL] = normalizeIPData(apiData, apiSource);
                                    }
                                }
                                publicipData.value = result;
                            }
                        } else {
                            showToast(labels.publicip_failed, 'error');
                        }
                    } catch (e) {
                        showToast(labels.publicip_failed + ': ' + e.message, 'error');
                    } finally {
                        publicipLoading.value = false;
                    }
                };

                const showServiceUpdate = function(id) {
                    updateTarget.value = id;
                    updateSelectedFile.value = null;
                    updateLoading.value = false;
                };

                const onUpdateFileSelected = function(event) {
                    updateSelectedFile.value = event.target.files[0];
                };

                const confirmServiceUpdate = async function() {
                    if (!updateTarget.value) {
                        showToast(labels.update_toast_select_client, 'error');
                        return;
                    }
                    if (!updateSelectedFile.value) {
                        showToast(labels.update_toast_select_file, 'error');
                        return;
                    }
                    updateLoading.value = true;
                    try {
                        const formData = new FormData();
                        formData.append('client_id', updateTarget.value);
                        formData.append('file', updateSelectedFile.value);
                        const data = await api('/api/service/update', {
                            method: 'POST',
                            body: formData
                        });
                        if (data.error) {
                            showToast(labels.update_failed.replace('{error}', data.error), 'error');
                        } else if (data.task_id) {
                            showToast(labels.update_success, 'success');
                            updateTarget.value = null;
                            updateSelectedFile.value = null;
                        }
                    } catch (e) {
                        showToast(labels.update_failed.replace('{error}', e.message), 'error');
                    } finally {
                        updateLoading.value = false;
                    }
                };

                const captureScreen = async function() {
                    if (!screencapTarget.value) {
                        showToast(labels.screencap_failed, 'error');
                        return;
                    }
                    screencapLoading.value = true;
                    try {
                        const data = await api('/api/screen/capture', {
                            method: 'POST',
                            body: JSON.stringify({
                                client_id: screencapTarget.value,
                                format: screencapFormat.value,
                                quality: screencapQuality.value,
                                display_index: screencapDisplay.value
                            })
                        });
                        if (data.error) {
                            showToast(labels.screencap_failed + ': ' + data.error, 'error');
                        } else if (data.response) {
                            var p = data.response;
                            if (p.error) {
                                showToast(labels.screencap_failed + ': ' + p.error, 'error');
                            } else {
                                screencapImageData.value = p.image_data || '';
                                screencapImageFormat.value = p.format || 'png';
                                screencapWidth.value = p.width || 0;
                                screencapHeight.value = p.height || 0;
                                screencapDisplayIndex.value = p.display_index || 0;
                                screencapDisplayCount.value = p.display_count || 0;
                                if (screencapAutoDownload.value && screencapImageData.value) {
                                    saveScreenshot();
                                }
                            }
                        } else {
                            showToast(labels.screencap_failed, 'error');
                        }
                    } catch (e) {
                        showToast(labels.screencap_failed + ': ' + e.message, 'error');
                    } finally {
                        screencapLoading.value = false;
                    }
                };

                const saveScreenshot = function() {
                    if (!screencapImageData.value) return;
                    try {
                        var link = document.createElement('a');
                        link.download = 'screenshot_' + Date.now() + '.' + screencapImageFormat.value;
                        link.href = 'data:image/' + screencapImageFormat.value + ';base64,' + screencapImageData.value;
                        link.click();
                        showToast(labels.screencap_save_success, 'success');
                    } catch (e) {
                        showToast(labels.screencap_save_failed, 'error');
                    }
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
                        'total': labels.sysinfo_table_total,
                        'used': labels.sysinfo_table_used,
                        'available': labels.sysinfo_na,
                        'free': labels.sysinfo_table_free,
                        'active': labels.sysinfo_na,
                        'inactive': labels.sysinfo_na,
                        'buffers': labels.sysinfo_na,
                        'cached': labels.sysinfo_na,
                        'sin': labels.sysinfo_na,
                        'sout': labels.sysinfo_na,
                        'used_percent': labels.sysinfo_na,
                        'hostname': labels.sysinfo_hostname,
                        'os': labels.sysinfo_os,
                        'platform': labels.sysinfo_platform,
                        'platform_version': labels.sysinfo_na,
                        'platform_family': labels.sysinfo_platform_family,
                        'kernel_version': labels.sysinfo_kernel_version,
                        'kernel_arch': labels.sysinfo_kernel_arch,
                        'uptime': labels.sysinfo_uptime,
                        'boot_time': labels.sysinfo_boot_time,
                        'procs': labels.sysinfo_processes_count_label,
                        'model_name': labels.sysinfo_model,
                        'cores': labels.sysinfo_cores,
                        'mhz': labels.sysinfo_mhz,
                        'cache_size': labels.sysinfo_cache_size,
                        'vendor_id': labels.sysinfo_vendor,
                        'read_count': labels.sysinfo_read_count,
                        'write_count': labels.sysinfo_write_count,
                        'read_bytes': labels.sysinfo_read_bytes,
                        'write_bytes': labels.sysinfo_write_bytes,
                        'read_time': labels.sysinfo_read_time,
                        'write_time': labels.sysinfo_write_time,
                        'bytes_sent': labels.sysinfo_bytes_sent,
                        'bytes_recv': labels.sysinfo_bytes_recv,
                        'packets_sent': labels.sysinfo_packets_sent,
                        'packets_recv': labels.sysinfo_packets_recv,
                        'errin': labels.sysinfo_errors_in,
                        'errout': labels.sysinfo_errors_out,
                        'dropin': labels.sysinfo_dropped_in,
                        'dropout': labels.sysinfo_dropped_out,
                        'mtu': labels.sysinfo_mtu,
                        'flags': labels.sysinfo_flags,
                        'addresses': labels.sysinfo_addresses
                    };
                    if (nameMap[key]) {
                        return nameMap[key];
                    }
                    return key.split('_').map(function(word) {
                        return word.charAt(0).toUpperCase() + word.slice(1);
                    }).join(' ');
                };

                const formatFieldValue = function(key, value) {
                    if (value === undefined || value === null) return labels.sysinfo_na;
                    
                    switch (key) {
                        case 'used_percent':
                            return typeof value === 'number' ? value.toFixed(1) + '%' : labels.sysinfo_na;
                        case 'uptime':
                            return formatUptime(value);
                        case 'boot_time':
                            return value ? new Date(value * 1000).toLocaleString() : labels.sysinfo_na;
                        case 'cache_size':
                            return typeof value === 'number' ? value + ' KB' : labels.sysinfo_na;
                        case 'mhz':
                            return typeof value === 'number' ? value.toFixed(2) : labels.sysinfo_na;
                        case 'read_time':
                        case 'write_time':
                            return typeof value === 'number' ? value + ' ms' : labels.sysinfo_na;
                        case 'flags':
                            return Array.isArray(value) ? value.join(', ') : (value || labels.sysinfo_na);
                        case 'addresses':
                            return Array.isArray(value) ? value.join(', ') : (value || labels.sysinfo_na);
                        default:
                            if (typeof value === 'number' && (key.includes('bytes') || key === 'total' || key === 'used' || key === 'free' || key === 'available' || key === 'active' || key === 'inactive' || key === 'buffers' || key === 'cached' || key === 'sin' || key === 'sout' || key.includes('sent') || key.includes('recv'))) {
                                return formatBytes(value);
                            }
                            return value || labels.sysinfo_na;
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
                    if (!unixTime) return labels.sysinfo_na;
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

                const showCopyFile = function(f) {
                    var sep = fmPath.value.includes('\\') ? '\\' : '/';
                    var fullPath = fmPath.value ? (fmPath.value.endsWith(sep) ? fmPath.value + f.name : fmPath.value + sep + f.name) : f.name;
                    fmCopyOrigin.value = fullPath;
                    fmCopyDest.value = '';
                    fmCopyTarget.value = true;
                };

                const confirmCopyFile = async function() {
                    if (!fmCopyDest.value.trim()) return;
                    var origin = fmCopyOrigin.value;
                    var dest = fmCopyDest.value.trim();
                    fmCopyTarget.value = false;

                    dest = resolvePath(fmPath.value, dest);

                    try {
                        var data = await api('/api/file/copy', {
                            method: 'POST',
                            body: JSON.stringify({ client_id: selectedClient.value, origin_path: origin, new_path: dest })
                        });
                        if (data.error) {
                            showToast(labels.fm_copy_failed.replace('{error}', data.error), 'error');
                        } else {
                            showToast(labels.fm_copy_success, 'success');
                            loadFiles();
                        }
                    } catch (e) {
                        showToast(labels.fm_copy_failed.replace('{error}', e.message), 'error');
                    }
                };

                const showProperties = function(f) {
                    fmPropertiesTarget.value = f;
                };

                const formatTransferProgress = function(current, total) {
                    return labels.fm_transfer_file_progress.replace('{current}', current).replace('{total}', total);
                };

                const triggerFileUpload = function() {
                    if (!selectedClient.value) {
                        showToast(labels.fm_toast_select_client, 'error');
                        return;
                    }
                    showUploadMenu.value = false;
                    document.getElementById('fileInput').click();
                };

                const triggerFolderUpload = function() {
                    if (!selectedClient.value) {
                        showToast(labels.fm_toast_select_client, 'error');
                        return;
                    }
                    showUploadMenu.value = false;
                    document.getElementById('folderInput').click();
                };

                const handleFileSelect = function(event) {
                    const files = event.target.files;
                    if (!files || files.length === 0) return;

                    const clientId = selectedClient.value;
                    const basePath = fmPath.value || '.';
                    let uploaded = 0;
                    let failed = 0;

                    const uploadNext = function(index) {
                        if (index >= files.length) {
                            event.target.value = '';
                            if (failed === 0 && uploaded > 0) {
                                showToast(labels.fm_upload_success, 'success');
                                loadFiles();
                            } else if (uploaded > 0) {
                                showToast(uploaded + ' uploaded, ' + failed + ' failed', 'warning');
                                loadFiles();
                            }
                            return;
                        }

                        const file = files[index];
                        const formData = new FormData();
                        formData.append('file', file);
                        formData.append('client_id', clientId);

                        let remotePath;
                        if (file.webkitRelativePath) {
                            const relPath = file.webkitRelativePath;
                            const slashIdx = relPath.indexOf('/');
                            if (slashIdx !== -1) {
                                const rootDirName = relPath.substring(0, slashIdx);
                                const relWithoutRoot = relPath.substring(slashIdx + 1);
                                remotePath = basePath === '.' ? rootDirName + '/' + relWithoutRoot : basePath + '/' + rootDirName + '/' + relWithoutRoot;
                            } else {
                                remotePath = basePath === '.' ? relPath : basePath + '/' + relPath;
                            }
                        } else {
                            remotePath = basePath === '.' ? file.name : basePath + '/' + file.name;
                        }
                        formData.append('remote_path', remotePath);

                        transferTask.value = {
                            type: 'upload',
                            fileName: file.name,
                            percent: 0,
                            fileIndex: index + 1,
                            fileCount: files.length
                        };

                        api('/api/file/upload', {
                            method: 'POST',
                            body: formData
                        }).then(function(resp) {
                            const taskId = resp.task_id;
                            listenTaskProgress(taskId, 'upload', function() {
                                uploaded++;
                                uploadNext(index + 1);
                            });
                        }).catch(function(err) {
                            failed++;
                            uploadNext(index + 1);
                        });
                    };

                    uploadNext(0);
                };

                const downloadFile = function(f) {
                    if (!selectedClient.value) {
                        showToast(labels.fm_toast_select_client, 'error');
                        return;
                    }

                    const remotePath = fmPath.value === '.' ? f.name : fmPath.value + '/' + f.name;

                    transferTask.value = {
                        type: 'download',
                        fileName: f.name,
                        percent: 0,
                        fileIndex: 1,
                        fileCount: 1
                    };

                    api('/api/file/download', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                            client_id: selectedClient.value,
                            remote_path: remotePath
                        })
                    }).then(function(resp) {
                        const taskId = resp.task_id;
                        listenTaskProgress(taskId, 'download', function() {
                            showToast(labels.fm_download_success, 'success');
                            loadFiles();
                        });
                    }).catch(function(err) {
                        transferTask.value = null;
                        showToast(labels.fm_download_failed.replace('{error}', err.message || 'Unknown'), 'error');
                    });
                };

                const listenTaskProgress = function(taskId, type, onDone) {
                    var pollInterval = null;
                    var pollCount = 0;
                    var maxPolls = 600;

                    var pollTask = function() {
                        pollCount++;
                        if (pollCount > maxPolls) {
                            clearInterval(pollInterval);
                            transferTask.value = null;
                            const msg = type === 'upload' ? labels.fm_upload_failed : labels.fm_download_failed;
                            showToast(msg.replace('{error}', 'Task timeout'), 'error');
                            return;
                        }

                        api('/api/task/status?task_id=' + taskId).then(function(data) {
                            if (data.status === 'processing' || data.status === 'pending') {
                                if (transferTask.value) {
                                    transferTask.value.percent = data.percent || 0;
                                    transferTask.value.fileName = data.file_name || transferTask.value.fileName;
                                    transferTask.value.fileIndex = data.file_index || transferTask.value.fileIndex;
                                    transferTask.value.fileCount = data.file_count || transferTask.value.fileCount;
                                }
                            } else if (data.status === 'done') {
                                clearInterval(pollInterval);
                                if (type === 'download') {
                                    window.location.href = basePath + '/api/file/download_result?task_id=' + taskId;
                                }
                                if (onDone) onDone();
                                setTimeout(function() {
                                    if (transferTask.value && transferTask.value.type === type) {
                                        transferTask.value = null;
                                    }
                                }, 2000);
                            } else if (data.status === 'error') {
                                clearInterval(pollInterval);
                                transferTask.value = null;
                                const msg = type === 'upload' ? labels.fm_upload_failed : labels.fm_download_failed;
                                showToast(msg.replace('{error}', data.error || 'Unknown error'), 'error');
                            }
                        }).catch(function(err) {
                            clearInterval(pollInterval);
                            transferTask.value = null;
                            const msg = type === 'upload' ? labels.fm_upload_failed : labels.fm_download_failed;
                            showToast(msg.replace('{error}', err.message || 'Connection lost'), 'error');
                        });
                    };

                    pollInterval = setInterval(pollTask, 500);
                    pollTask();
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
                    fmCopyTarget: fmCopyTarget,
                    fmCopyOrigin: fmCopyOrigin,
                    fmCopyDest: fmCopyDest,
                    fmPropertiesTarget: fmPropertiesTarget,
                    transferTask: transferTask,
                    showUploadMenu: showUploadMenu,
                    triggerFileUpload: triggerFileUpload,
                    triggerFolderUpload: triggerFolderUpload,
                    handleFileSelect: handleFileSelect,
                    downloadFile: downloadFile,
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
                    showCopyFile: showCopyFile,
                    confirmCopyFile: confirmCopyFile,
                    showProperties: showProperties,
                    formatTransferProgress: formatTransferProgress,
                    screencapTarget: screencapTarget,
                    screencapFormat: screencapFormat,
                    screencapQuality: screencapQuality,
                    screencapDisplay: screencapDisplay,
                    screencapLoading: screencapLoading,
                    screencapImageData: screencapImageData,
                    screencapImageFormat: screencapImageFormat,
                    screencapWidth: screencapWidth,
                    screencapHeight: screencapHeight,
                    screencapDisplayIndex: screencapDisplayIndex,
                    screencapDisplayCount: screencapDisplayCount,
                    screencapAutoDownload: screencapAutoDownload,
                    showScreenCapture: showScreenCapture,
                    captureScreen: captureScreen,
                    saveScreenshot: saveScreenshot,
                    publicipTarget: publicipTarget,
                    publicipData: publicipData,
                    publicipLoading: publicipLoading,
                    showPublicIP: showPublicIP,
                    fetchPublicIP: fetchPublicIP,
                    updateTarget: updateTarget,
                    updateSelectedFile: updateSelectedFile,
                    updateLoading: updateLoading,
                    showServiceUpdate: showServiceUpdate,
                    onUpdateFileSelected: onUpdateFileSelected,
                    confirmServiceUpdate: confirmServiceUpdate,
                    moreMenuId: moreMenuId,
                    moreMenuTop: moreMenuTop,
                    moreMenuLeft: moreMenuLeft,
                    toggleMoreMenu: toggleMoreMenu
                };
            }
        }).mount('#app');