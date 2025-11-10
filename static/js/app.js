class FlightSearchApp {
    constructor() {
        this.currentTab = 'search';
        this.initEventListeners();
        this.setDefaultDates();
        this.showTab('search');
        this.initCurrencyCalculator(); // 初始化匯率計算機
    }

    initEventListeners() {
        const searchForm = document.getElementById('searchForm');
        const trackingForm = document.getElementById('trackingForm');
        
        searchForm.addEventListener('submit', (e) => this.handleSearch(e));
        trackingForm.addEventListener('submit', (e) => this.handleTracking(e));

        // 機場自動完成
        this.setupAutocomplete('origin');
        this.setupAutocomplete('destination');
        this.setupAutocomplete('trackingOrigin');
        this.setupAutocomplete('trackingDestination');

        // 標籤切換
        document.querySelectorAll('.nav-tab').forEach(tab => {
            tab.addEventListener('click', (e) => {
                const tabName = e.target.dataset.tab;
                this.showTab(tabName);
            });
        });
    }

    showTab(tabName) {
        // 更新活躍標籤
        document.querySelectorAll('.nav-tab').forEach(tab => {
            tab.classList.toggle('active', tab.dataset.tab === tabName);
        });

        // 顯示對應內容
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.toggle('active', content.id === `${tabName}Tab`);
        });

        this.currentTab = tabName;
        
        // 隱藏所有結果區域
        this.hideElement('results');
        this.hideElement('trackingResults');
        this.hideElement('error');
        this.hideElement('trackingError');
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
        if (!input) return;

        const suggestions = document.getElementById(fieldId + 'Suggestions');
        if (!suggestions) return;

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

    // 即時航班搜尋
    async handleSearch(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const params = new URLSearchParams(formData);
        
        console.log('🔍 發送搜尋請求:', params.toString());
        
        this.hideElement('results');
        this.hideElement('error');
        this.showElement('loading');

        try {
            const response = await fetch(`/api/flights/search?${params}`);
            const data = await response.json();

            console.log('📊 收到搜尋響應:', data);

            if (!response.ok) {
                throw new Error(data.error || '搜尋失敗');
            }

            this.displayResults(data);
        } catch (error) {
            console.error('❌ 搜尋錯誤:', error);
            this.showError(error.message);
        } finally {
            this.hideElement('loading');
        }
    }

    // 價格追蹤功能
    async handleTracking(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const origin = formData.get('trackingOrigin');
        const destination = formData.get('trackingDestination');
        const weeks = formData.get('weeks') || 12;
        
        this.hideElement('trackingResults');
        this.hideElement('trackingError');
        this.showElement('trackingLoading');

        try {
            const params = new URLSearchParams({
                origin,
                destination,
                weeks
            });

            console.log('🔍 發送價格追蹤請求:', params.toString());
            
            const response = await fetch(`/api/flights/track-prices?${params}`);
            const data = await response.json();

            console.log('📊 收到價格追蹤響應:', data);

            if (!response.ok) {
                throw new Error(data.error || '價格追蹤失敗');
            }

            this.displayTrackingResults(data);
        } catch (error) {
            console.error('❌ 價格追蹤錯誤:', error);
            this.showTrackingError(error.message);
        } finally {
            this.hideElement('trackingLoading');
        }
    }

    displayResults(data) {
        console.log('🎯 開始顯示結果:', data);
        
        const resultsDiv = document.getElementById('results');
        const countDiv = document.getElementById('resultsCount');
        const flightsDiv = document.getElementById('flightsList');
        const weatherDiv = document.getElementById('weatherInfo');
        const exchangeDiv = document.getElementById('exchangeInfo');

        // 檢查元素是否存在
        if (!resultsDiv || !countDiv || !flightsDiv || !weatherDiv || !exchangeDiv) {
            console.error('❌ 找不到必要的DOM元素');
            return;
        }

        // 清空之前的結果
        flightsDiv.innerHTML = '';

        if (!data.success) {
            flightsDiv.innerHTML = `
                <div style="text-align: center; padding: 40px; color: #dc3545;">
                    <i class="fas fa-exclamation-triangle" style="font-size: 3rem; margin-bottom: 20px;"></i>
                    <h3>搜尋失敗</h3>
                    <p>${data.error || '未知錯誤'}</p>
                </div>
            `;
            this.hideElement('weatherInfo');
            this.hideElement('exchangeInfo');
        } else if (!data.data || !data.data.flights || data.data.flights.length === 0) {
            countDiv.textContent = '找到 0 個航班';
            flightsDiv.innerHTML = `
                <div style="text-align: center; padding: 40px; color: #666;">
                    <i class="fas fa-plane-slash" style="font-size: 3rem; margin-bottom: 20px;"></i>
                    <h3>沒有找到符合條件的航班</h3>
                    <p>請嘗試調整搜尋條件</p>
                </div>
            `;
            this.hideElement('weatherInfo');
            this.hideElement('exchangeInfo');
        } else {
            const flights = data.data.flights;
            const weatherInfo = data.data.weather;
            const exchangeInfo = data.data.exchange;
            
            countDiv.textContent = `找到 ${data.data.meta?.count || flights.length} 個航班`;
            console.log(`📈 顯示 ${flights.length} 個航班`);
            
            // 顯示天氣資訊
            if (weatherInfo) {
                this.displayWeatherInfo(weatherInfo);
                this.showElement('weatherInfo');
            } else {
                this.hideElement('weatherInfo');
            }

            // 顯示匯率資訊
            if (exchangeInfo) {
                this.displayExchangeInfo(exchangeInfo);
                this.showElement('exchangeInfo');
            } else {
                this.hideElement('exchangeInfo');
            }
            
            // 顯示航班列表
            flights.forEach((flight, index) => {
                console.log(`✈️ 航班 ${index + 1}:`, flight);
                try {
                    const flightCard = this.createFlightCard(flight);
                    flightsDiv.innerHTML += flightCard;
                } catch (error) {
                    console.error(`❌ 創建航班卡片 ${index + 1} 失敗:`, error);
                }
            });
        }

        this.showElement('results');
        console.log('✅ 結果顯示完成');
    }

    // 顯示天氣資訊
    displayWeatherInfo(weatherInfo) {
        console.log('🌤️ 顯示天氣資訊:', weatherInfo);
        
        // 出發地天氣
        if (weatherInfo.origin_weather) {
            const origin = weatherInfo.origin_weather;
            document.getElementById('originTemp').textContent = `${Math.round(origin.avg_temp)}°C`;
            document.getElementById('originCondition').textContent = origin.condition;
            document.getElementById('originHumidity').textContent = origin.humidity;
            document.getElementById('originWind').textContent = origin.wind_speed;
            document.getElementById('originRain').textContent = origin.chance_of_rain || 0;
            
            // 更新城市名稱顯示
            const originCityElement = document.querySelector('#originWeather h4');
            if (originCityElement) {
                originCityElement.innerHTML = `<i class="fas fa-plane-departure"></i> ${origin.city} 天氣`;
            }
            
            // 設定天氣圖標
            const originIcon = document.getElementById('originWeatherIcon');
            if (origin.icon && originIcon) {
                originIcon.src = `https:${origin.icon}`;
                originIcon.alt = origin.condition;
            }
        }

        // 目的地天氣
        if (weatherInfo.destination_weather) {
            const destination = weatherInfo.destination_weather;
            document.getElementById('destinationTemp').textContent = `${Math.round(destination.avg_temp)}°C`;
            document.getElementById('destinationCondition').textContent = destination.condition;
            document.getElementById('destinationHumidity').textContent = destination.humidity;
            document.getElementById('destinationWind').textContent = destination.wind_speed;
            document.getElementById('destinationRain').textContent = destination.chance_of_rain || 0;
            
            // 更新城市名稱顯示
            const destinationCityElement = document.querySelector('#destinationWeather h4');
            if (destinationCityElement) {
                destinationCityElement.innerHTML = `<i class="fas fa-plane-arrival"></i> ${destination.city} 天氣`;
            }
            
            // 設定天氣圖標
            const destinationIcon = document.getElementById('destinationWeatherIcon');
            if (destination.icon && destinationIcon) {
                destinationIcon.src = `https:${destination.icon}`;
                destinationIcon.alt = destination.condition;
            }
        }

        // 旅行建議
        if (weatherInfo.travel_advice) {
            document.getElementById('adviceText').textContent = weatherInfo.travel_advice;
        }
    }

    // 顯示匯率資訊
    displayExchangeInfo(exchangeInfo) {
        console.log('💱 顯示匯率資訊:', exchangeInfo);
        
        // 顯示基礎貨幣
        document.getElementById('baseCurrency').textContent = exchangeInfo.base_currency;
        
        // 顯示更新時間
        const lastUpdated = new Date(exchangeInfo.last_updated).toLocaleString('zh-TW');
        document.getElementById('exchangeLastUpdated').textContent = lastUpdated;
        
        // 顯示匯率卡片
        const ratesContainer = document.getElementById('exchangeRates');
        ratesContainer.innerHTML = '';
        
        Object.entries(exchangeInfo.rates).forEach(([currency, rate]) => {
            const currencyNames = {
                'USD': '美元', 'EUR': '歐元', 'JPY': '日圓', 'GBP': '英鎊',
                'CNY': '人民幣', 'KRW': '韓元', 'HKD': '港幣', 'SGD': '新加坡元',
                'TWD': '新台幣'
            };
            
            const rateCard = document.createElement('div');
            rateCard.className = 'exchange-rate-card';
            rateCard.innerHTML = `
                <div class="currency-code">${currency}</div>
                <div class="currency-rate">${rate.toFixed(4)}</div>
                <div class="currency-name">${currencyNames[currency] || currency}</div>
            `;
            ratesContainer.appendChild(rateCard);
        });
        
        // 初始化貨幣轉換工具
        this.initCurrencyConverter();
    }

    // 初始化貨幣轉換工具
    initCurrencyConverter() {
        const convertBtn = document.getElementById('convertBtn');
        if (convertBtn) {
            convertBtn.addEventListener('click', () => this.handleCurrencyConversion());
        }
        
        // 也支援 Enter 鍵轉換
        const amountInput = document.getElementById('convertAmount');
        if (amountInput) {
            amountInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.handleCurrencyConversion();
                }
            });
        }
    }

    // 處理貨幣轉換
    async handleCurrencyConversion() {
        const amount = parseFloat(document.getElementById('convertAmount').value);
        const fromCurrency = document.getElementById('convertFrom').value;
        const toCurrency = document.getElementById('convertTo').value;
        const resultDiv = document.getElementById('conversionResult');
        
        if (!amount || amount <= 0) {
            resultDiv.innerHTML = '<span style="color: #ff6b6b;">請輸入有效的金額</span>';
            return;
        }
        
        if (fromCurrency === toCurrency) {
            resultDiv.innerHTML = '<span style="color: #ff6b6b;">請選擇不同的貨幣</span>';
            return;
        }
        
        try {
            const response = await fetch('/api/currency/convert', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    amount: amount,
                    from_currency: fromCurrency,
                    to_currency: toCurrency
                })
            });
            
            const data = await response.json();
            
            if (data.success) {
                const result = data.data;
                resultDiv.innerHTML = `
                    <div style="font-size: 1.1em;">
                        ${this.formatPrice(result.original_amount)} ${result.from_currency} = 
                        <span style="color: #ffeb3b; font-size: 1.2em;">
                            ${this.formatPrice(result.converted_amount)} ${result.to_currency}
                        </span>
                    </div>
                    <div style="font-size: 0.9em; opacity: 0.8; margin-top: 5px;">
                        匯率: 1 ${result.from_currency} = ${result.exchange_rate.toFixed(4)} ${result.to_currency}
                    </div>
                `;
            } else {
                resultDiv.innerHTML = `<span style="color: #ff6b6b;">轉換失敗: ${data.error}</span>`;
            }
        } catch (error) {
            console.error('貨幣轉換錯誤:', error);
            resultDiv.innerHTML = '<span style="color: #ff6b6b;">轉換服務暫時不可用</span>';
        }
    }

    // 新增：初始化匯率計算機
    initCurrencyCalculator() {
        const calculateBtn = document.getElementById('calculateBtn');
        const swapBtn = document.getElementById('swapCurrencies');
        const amountInput = document.getElementById('calcAmount');

        if (calculateBtn) {
            calculateBtn.addEventListener('click', () => this.handleCurrencyCalculation());
        }

        if (swapBtn) {
            swapBtn.addEventListener('click', () => this.swapCurrencies());
        }

        if (amountInput) {
            amountInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.handleCurrencyCalculation();
                }
            });

            // 實時計算
            amountInput.addEventListener('input', () => {
                if (amountInput.value) {
                    this.handleCurrencyCalculation();
                }
            });
        }

        // 當貨幣選擇改變時自動計算
        const fromSelect = document.getElementById('calcFromCurrency');
        const toSelect = document.getElementById('calcToCurrency');
        
        if (fromSelect) {
            fromSelect.addEventListener('change', () => {
                if (amountInput.value) {
                    this.handleCurrencyCalculation();
                }
            });
        }

        if (toSelect) {
            toSelect.addEventListener('change', () => {
                if (amountInput.value) {
                    this.handleCurrencyCalculation();
                }
            });
        }

        // 載入時顯示即時匯率
        this.loadLiveRates();
    }

    // 新增：交換貨幣
    swapCurrencies() {
        const fromSelect = document.getElementById('calcFromCurrency');
        const toSelect = document.getElementById('calcToCurrency');
        
        const fromValue = fromSelect.value;
        const toValue = toSelect.value;
        
        fromSelect.value = toValue;
        toSelect.value = fromValue;
        
        // 如果金額不為空，重新計算
        const amountInput = document.getElementById('calcAmount');
        if (amountInput.value) {
            this.handleCurrencyCalculation();
        }
    }

    // 新增：處理貨幣計算
    async handleCurrencyCalculation() {
        const amount = parseFloat(document.getElementById('calcAmount').value);
        const fromCurrency = document.getElementById('calcFromCurrency').value;
        const toCurrency = document.getElementById('calcToCurrency').value;
        const resultDiv = document.getElementById('calcResult');
        
        if (!amount || amount <= 0) {
            this.hideElement('calcResult');
            return;
        }
        
        if (fromCurrency === toCurrency) {
            resultDiv.innerHTML = `
                <div style="text-align: center; padding: 20px; color: #666;">
                    <i class="fas fa-info-circle" style="font-size: 2rem; margin-bottom: 10px;"></i>
                    <p>請選擇不同的貨幣進行轉換</p>
                </div>
            `;
            this.showElement('calcResult');
            return;
        }
        
        try {
            const response = await fetch('/api/currency/convert', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    amount: amount,
                    from_currency: fromCurrency,
                    to_currency: toCurrency
                })
            });
            
            const data = await response.json();
            
            if (data.success) {
                const result = data.data;
                this.displayCalculationResult(result);
                this.showElement('calcResult');
            } else {
                this.showCalculationError(data.error || '轉換失敗');
            }
        } catch (error) {
            console.error('貨幣計算錯誤:', error);
            this.showCalculationError('轉換服務暫時不可用');
        }
    }

    // 新增：顯示計算結果
    displayCalculationResult(result) {
        document.getElementById('originalAmountDisplay').textContent = 
            this.formatPrice(result.original_amount);
        document.getElementById('fromCurrencyDisplay').textContent = result.from_currency;
        document.getElementById('convertedAmountDisplay').textContent = 
            this.formatPrice(result.converted_amount);
        document.getElementById('toCurrencyDisplay').textContent = result.to_currency;
        
        // 顯示匯率
        document.getElementById('exchangeRateDisplay').textContent = 
            `1 ${result.from_currency} = ${result.exchange_rate.toFixed(6)} ${result.to_currency}`;
        
        // 顯示反向匯率
        const reverseRate = 1 / result.exchange_rate;
        document.getElementById('reverseRateDisplay').textContent = 
            `1 ${result.to_currency} = ${reverseRate.toFixed(6)} ${result.from_currency}`;
        
        // 顯示更新時間
        const lastUpdated = new Date(result.last_updated).toLocaleString('zh-TW');
        document.getElementById('calcLastUpdated').textContent = lastUpdated;
    }

    // 新增：顯示計算錯誤
    showCalculationError(message) {
        const resultDiv = document.getElementById('calcResult');
        resultDiv.innerHTML = `
            <div style="text-align: center; padding: 20px; color: #dc3545;">
                <i class="fas fa-exclamation-triangle" style="font-size: 2rem; margin-bottom: 10px;"></i>
                <p>${message}</p>
            </div>
        `;
        this.showElement('calcResult');
    }

    // 新增：載入即時匯率
    async loadLiveRates() {
        try {
            const baseCurrency = 'TWD'; // 使用 TWD 作為基礎貨幣
            const targetCurrencies = ['USD', 'EUR', 'JPY', 'GBP', 'CNY', 'KRW', 'HKD', 'SGD'];
            
            // 這裡可以呼叫 API 獲取即時匯率並顯示在表格中
            console.log('載入即時匯率資料...');
            
        } catch (error) {
            console.error('載入即時匯率失敗:', error);
        }
    }

    displayTrackingResults(data) {
        const resultsDiv = document.getElementById('trackingResults');
        const analysis = data.data;
        
        console.log('📈 顯示價格分析:', analysis);
        
        resultsDiv.innerHTML = this.createTrackingAnalysis(analysis);
        this.showElement('trackingResults');
    }

    createTrackingAnalysis(analysis) {
        const bestDate = new Date(analysis.best_date).toLocaleDateString('zh-TW');
        
        return `
            <div class="analysis-summary">
                <h3 style="margin-bottom: 20px; text-align: center;"><i class="fas fa-chart-bar"></i> 價格分析摘要</h3>
                <div class="summary-grid">
                    <div class="summary-item">
                        <h3>最低價格</h3>
                        <div class="value">$${this.formatPrice(analysis.min_price)}</div>
                    </div>
                    <div class="summary-item">
                        <h3>平均價格</h3>
                        <div class="value">$${this.formatPrice(analysis.avg_price)}</div>
                    </div>
                    <div class="summary-item">
                        <h3>最高價格</h3>
                        <div class="value">$${this.formatPrice(analysis.max_price)}</div>
                    </div>
                    <div class="summary-item">
                        <h3>最佳出發</h3>
                        <div class="value">${bestDate}</div>
                    </div>
                </div>
                <div class="recommendation" style="background: rgba(255, 255, 255, 0.2); padding: 15px; border-radius: 8px; margin-top: 15px; text-align: center;">
                    <strong>💡 ${analysis.recommendation || '建議根據價格趨勢選擇出發時間'}</strong>
                </div>
            </div>

            <div style="background: #f8f9fa; border-radius: 10px; padding: 20px; margin: 20px 0;">
                <h4><i class="fas fa-history"></i> 價格時間軸</h4>
                <div style="max-height: 400px; overflow-y: auto;">
                    ${analysis.data_points ? analysis.data_points.map(point => this.createTimelineItem(point)).join('') : '沒有價格數據'}
                </div>
            </div>

            <div style="background: #e8f5e8; border: 1px solid #c3e6cb; border-radius: 8px; padding: 15px; color: #155724;">
                <i class="fas fa-check-circle"></i>
                <strong> 分析完成！</strong>
                <p style="margin: 5px 0 0 0;">已分析 ${analysis.track_weeks} 週的價格數據，建議您在 ${bestDate} 附近出發可獲得最優價格。</p>
            </div>
        `;
    }

    createTimelineItem(point) {
        const date = new Date(point.date).toLocaleDateString('zh-TW');
        const isBestPrice = point.price === point.min_price;
        
        return `
            <div style="display: flex; justify-content: space-between; align-items: center; padding: 12px; background: white; border-radius: 8px; margin: 8px 0; border-left: 4px solid ${isBestPrice ? '#28a745' : '#667eea'};">
                <div>
                    <div style="font-weight: 600; color: #333;">${date}</div>
                    <div style="color: #666; font-size: 0.9rem;">第 ${point.week} 週</div>
                </div>
                <div style="font-size: 1.2rem; font-weight: 700; color: #e74c3c;">
                    $${this.formatPrice(point.price)}
                </div>
            </div>
        `;
    }

    createFlightCard(flight) {
        console.log('🎫 創建航班卡片:', flight);
        
        const departureTime = flight.departure ? new Date(flight.departure).toLocaleTimeString('zh-TW', { 
            hour: '2-digit', minute: '2-digit' 
        }) : '未知';
        
        const arrivalTime = flight.arrival ? new Date(flight.arrival).toLocaleTimeString('zh-TW', { 
            hour: '2-digit', minute: '2-digit' 
        }) : '未知';

        const airline = flight.airline || '未知航空公司';
        const stops = flight.stops || 0;
        const price = flight.price || 0;
        const currency = flight.currency || 'TWD';

        return `
            <div class="flight-card">
                <div class="flight-info">
                    <div class="flight-route">
                        <div class="flight-airports">
                            ${flight.from?.code || '未知'} → ${flight.to?.code || '未知'}
                        </div>
                        <div class="flight-duration">
                            ${this.formatDuration(flight.duration)}
                        </div>
                    </div>
                    <div class="flight-details">
                        <span><i class="fas fa-plane"></i> ${airline}</span>
                        <span><i class="fas fa-clock"></i> ${departureTime} - ${arrivalTime}</span>
                        <span><i class="fas fa-stopwatch"></i> ${stops} 次停靠</span>
                        ${flight.flightNumber ? `<span><i class="fas fa-ticket-alt"></i> ${flight.flightNumber}</span>` : ''}
                    </div>
                </div>
                <div class="flight-price">
                    <div class="price">${this.formatPrice(price)}</div>
                    <div class="currency">${currency}</div>
                </div>
            </div>
        `;
    }

    formatDuration(duration) {
        if (!duration) return '未知時長';
        
        const match = duration.match(/PT(?:(\d+)H)?(?:(\d+)M)?/);
        if (!match) return duration;
        
        const hours = match[1] ? parseInt(match[1]) : 0;
        const minutes = match[2] ? parseInt(match[2]) : 0;
        
        let result = '';
        if (hours > 0) result += `${hours}小時`;
        if (minutes > 0) result += `${minutes}分鐘`;
        return result || '0分鐘';
    }

    formatPrice(price) {
        if (!price) return '0';
        return new Intl.NumberFormat('zh-TW').format(Math.round(price));
    }

    showError(message) {
        const errorDiv = document.getElementById('error');
        const errorMessage = document.getElementById('errorMessage');
        
        errorMessage.textContent = message;
        this.showElement('error');
    }

    showTrackingError(message) {
        const errorDiv = document.getElementById('trackingError');
        const errorMessage = document.getElementById('trackingErrorMessage');
        
        errorMessage.textContent = message;
        this.showElement('trackingError');
    }

    showElement(id) {
        const element = document.getElementById(id);
        if (element) {
            element.classList.remove('hidden');
        }
    }

    hideElement(id) {
        const element = document.getElementById(id);
        if (element) {
            element.classList.add('hidden');
        }
    }
}

// 初始化應用
document.addEventListener('DOMContentLoaded', () => {
    new FlightSearchApp();
    console.log('🚀 航班搜尋應用已初始化');
});