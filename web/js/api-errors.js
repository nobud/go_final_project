function apiErrorMessage(error) {
    const response = error && error.response;
    if (!response) {
        return "Не удалось связаться с сервером. Повторите попытку.";
    }
    if (response.status === 401) {
        return "401: требуется авторизация";
    }
    if (response.data && typeof response.data.error === "string" && response.data.error) {
        return response.data.error;
    }
    return "Ошибка запроса к серверу (HTTP " + response.status + ").";
}
