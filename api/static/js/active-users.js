$(document).ready(function () {
    var showData = $('#active-users');
    $.getJSON('https://www.syncmysport.com/users/count', function (data) {
      console.log(data);
      showData.text("Active users: " + data["active-users"]);
    });
});
