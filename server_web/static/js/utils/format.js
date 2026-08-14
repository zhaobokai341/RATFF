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
const formatModTime = function(unixTime) {
    if (!unixTime) return 'N/A';
    var d = new Date(unixTime * 1000);
    return d.toLocaleString();
};
