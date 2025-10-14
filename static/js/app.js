class FlightSearchApp {
    constructor() {
        this.initEventListeners();
        this.setDefaultDates();
    }

    initEventListeners() {
        const form = document.getElementById('searchForm');
        form.addEventListener('submit', (e) => this.handleSearch(e));

        // 機場自動完成
        this.setupAutocomplete('origin');
        this.setupAutocomplete('destination');
    }

    setDefaultDates() {
        const today = new Date().toISOString().split('T')[0];
        const nextWeek = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
        
        document.getElementById('departureDate').value = nextWeek;
        document.getElementById('departureDate').min = today;
        document.getElementById('returnDate').min = today;
    }

    setupAutocomplete(fieldId) {
        const input = document.getElementById(fieldId);
        const suggestions = document.getElementById(fieldId + 'Suggestions');

        input.addEventListener('input', (e) => {
            const query = e.target.value.trim();
            if (query.length >= 2) {
                this.searchAirports(query, suggestions);
            } else {
                suggestions.style.display = 'none';
            }
        });

        input.addEventListener('focus', () => {
            if (suggestions.children.length > 0) {
                suggestions.style.display = 'block';
            }
        });

        // 點擊外部隱藏建議
        document.addEventListener('click', (e) => {
            if (!input.contains(e.target) && !suggestions.contains(e.target)) {
                suggestions.style.display = 'none';
            }
        });
    }

    async searchAirports(query, suggestionsContainer) {
        try {
            const response = await fetch(`/api/airports/search?q=${encodeURIComponent(query)}`);
            const data = await response.json();

            if (data.success && data.data.length > 0) {
                this.showSuggestions(data.data, suggestionsContainer);
            } else {
                suggestionsContainer.style.display = 'none';
            }
        } catch (error) {
            console.error('搜尋機場失敗:', error);
        }
    }

    showSuggestions(airports, container) {
        container.innerHTML = '';
        airports.forEach(airport => {
            const div = document.createElement('div');
            div.className = 'suggestion-item';
            div.textContent = `${airport.code} - ${airport.name} (${airport.city})`;
            div.addEventListener('click', () => {
                const input = container.previousElementSibling;
                input.value = airport.code;
                container.style.display = 'none';
            });
            container.appendChild(div);
        });
        container.style.display = 'block';
    }

    async handleSearch(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const params = new URLSearchParams(formData);
        
        // 隱藏之前的結果和錯誤
        this.hideElement('results');
        this.hideElement('error');
        
        // 顯示載入動畫
        this.showElement('loading');

        try {
            const response = await fetch(`/api/flights/search?${params}`);
            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || '搜尋失敗');
            }

            this.displayResults(data);
        } catch (error) {
            this.showError(error.message);
        } finally {
            this.hideElement('loading');
        }
    }

    displayResults(data) {
        const resultsDiv = document.getElementById('results');
        const countDiv = document.getElementById('resultsCount');
        const flightsDiv = document.getElementById('flightsList');

        countDiv.textContent = `找到 ${data.count} 個航班`;
        
        if (data.count === 0) {
            flightsDiv.innerHTML = '<p class="no-results">沒有找到符合條件的航班</p>';
        } else {
            flightsDiv.innerHTML = data.data.map(flight => this.createFlightCard(flight)).join('');
        }

        this.showElement('results');
    }

    createFlightCard(flight) {
        const departureTime = new Date(flight.departure).toLocaleTimeString('zh-TW', { 
            hour: '2-digit', minute: '2-digit' 
        });
        const arrivalTime = new Date(flight.arrival).toLocaleTimeString('zh-TW', { 
            hour: '2-digit', minute: '2-digit' 
        });

        return `
            <div class="flight-card">
                <div class="flight-info">
                    <div class="flight-route">
                        <div class="flight-airports">
                            ${flight.from.code} → ${flight.to.code}
                        </div>
                        <div class="flight-duration">
                            ${this.formatDuration(flight.duration)}
                        </div>
                    </div>
                    <div class="flight-details">
                        <span><i class="fas fa-plane"></i> ${flight.airline}</span>
                        <span><i class="fas fa-clock"></i> ${departureTime} - ${arrivalTime}</span>
                        <span><i class="fas fa-stopwatch"></i> ${flight.stops} 次停靠</span>
                    </div>
                </div>
                <div class="flight-price">
                    <div class="price">${this.formatPrice(flight.price)}</div>
                    <div class="currency">${flight.currency}</div>
                </div>
            </div>
        `;
    }

    formatDuration(duration) {
        // 將 PT2H30M 轉換為 2小時30分鐘
        const match = duration.match(/PT(?:(\d+)H)?(?:(\d+)M)?/);
        const hours = match[1] ? parseInt(match[1]) : 0;
        const minutes = match[2] ? parseInt(match[2]) : 0;
        
        let result = '';
        if (hours > 0) result += `${hours}小時`;
        if (minutes > 0) result += `${minutes}分鐘`;
        return result || '0分鐘';
    }

    formatPrice(price) {
        return new Intl.NumberFormat('zh-TW').format(price);
    }

    showError(message) {
        const errorDiv = document.getElementById('error');
        const errorMessage = document.getElementById('errorMessage');
        
        errorMessage.textContent = message;
        this.showElement('error');
    }

    showElement(id) {
        document.getElementById(id).classList.remove('hidden');
    }

    hideElement(id) {
        document.getElementById(id).classList.add('hidden');
    }
}

// 初始化應用
document.addEventListener('DOMContentLoaded', () => {
    new FlightSearchApp();
});