

/**
 * // 使用示例
 * const apiClient = new HttpClient('https://api.example.com');
 *
 * // 发起GET请求
 * apiClient.get('/users')
 *     .then(data => console.log(data))
 *     .catch(error => console.error('Error fetching users:', error));
 *
 * // 发起POST请求
 * apiClient.post('/users', { name: 'John Doe', age: 30 })
 *     .then(data => console.log(data))
 *     .catch(error => console.error('Error creating user:', error));
 */
class HttpClient {
    constructor(baseUrl = '') {
        this.baseUrl = baseUrl;
    }

    async get(endpoint, params = {}) {
        const queryString = this._buildQueryString(params);
        const url = `${this.baseUrl}${endpoint}${queryString ? '?' + queryString : ''}`;
        const response = await fetch(url);
        return this._handleResponse(response);
    }

    async post(endpoint, data = {}) {
        const url = `${this.baseUrl}${endpoint}`;
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });
        return this._handleResponse(response);
    }

    // 可以根据需要添加其他HTTP方法，如put, delete等

    _buildQueryString(params) {
        return Object.keys(params)
            .map(key => `${encodeURIComponent(key)}=${encodeURIComponent(params[key])}`)
            .join('&');
    }

    async _handleResponse(response) {
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        return data;
    }

    async fetchFileAsByteArray(url) {
        try {
            // ʹ�� fetch API ��ȡ�ļ�����
            const response = await fetch(url);

            // �����Ӧ�Ƿ�ɹ�
            if (!response.ok) {
                throw new Error(`Failed to fetch file: ${response.statusText}`);
            }

            // ʹ�� ArrayBuffer �� Uint8Array ����ȡ�ֽ�����
            const arrayBuffer = await response.arrayBuffer();
            const byteArray = new Uint8Array(arrayBuffer);

            // �����ֽ�����
            return byteArray;
        } catch (error) {
            console.error('Error fetching and parsing file:', error);
        }
    }
}



