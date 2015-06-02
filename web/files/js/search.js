$(document).ready(function() {
    // apply plugin to allow results table to be sorted
    var table = $("#resultsTable");
    table.tablesorter();
});

$(function() {
    $("#reportSelected").button({
        icons: {
            primary: "ui-icon-locked"
        }

    }).click(function(evt) {

        var selected = [];
        console.log("Looking for checked boxes");
        $(".selectItemBox :checkbox:checked").each(function() {
            
        });
    });

});