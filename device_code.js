function copyFunc() {
    const input = document.getElementById("device_code");

    input.select();
    input.setSelectionRange(0, 99999);
    navigator.clipboard.writeText(input.value);
}

window.onload = function () {
    const queryParams = new URLSearchParams(window.location.search);
    const input = document.getElementById("device_code");
    input.value = queryParams.get("code");
}
