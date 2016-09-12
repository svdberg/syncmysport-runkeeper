$(document).ready(function () {
    var showData = $('#active-users');
    $.getJSON('http://www.syncmysport.com/users/count', function (data) {
      console.log(data);
      showData.text("Active users: " + data["active-users"]);
    });
});
