// LangSwitcher - Reusable Vue component for language switching
// Usage: <lang-switcher></lang-switcher>

const LangSwitcher = {
    template: `
        <div class="lang-switcher">
            <select 
                v-model="currentLang" 
                @change="onLangChange" 
                class="px-2 py-1 text-xs bg-transparent border border-border rounded text-gray-400 hover:text-gray-200 hover:bg-surface-2 transition cursor-pointer appearance-none"
                style="background-image: url(&quot;data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' viewBox='0 0 12 12'%3E%3Cpath fill='%23888' d='M6 8L1 3h10z'/%3E%3C/svg%3E&quot;); background-repeat: no-repeat; background-position: right 8px center; padding-right: 22px;"
            >
                <option value="zh">中文</option>
                <option value="en">English</option>
            </select>
        </div>
    `,
    data() {
        return {
            currentLang: 'zh'
        }
    },
    mounted() {
        this.currentLang = this.getCookie('app_lang') || 'zh'
    },
    methods: {
        onLangChange() {
            this.setCookie('app_lang', this.currentLang, 365)
            window.location.reload()
        },
        getCookie(name) {
            var match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'))
            return match ? decodeURIComponent(match[2]) : null
        },
        setCookie(name, value, days) {
            var expires = ''
            if (days) {
                var date = new Date()
                date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000))
                expires = '; expires=' + date.toUTCString()
            }
            document.cookie = name + '=' + encodeURIComponent(value) + expires + '; path=/'
        }
    }
}