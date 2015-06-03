$(document).ready(function() {
    // apply plugin to allow results table to be sorted
    var table = $("#resultsTable");
    table.tablesorter();
});

function findSelected(){
	var selected = [];
        console.log("Looking for checked boxes");

        $(".itemRow").each(function() {
            console.log("Have a row")
						
			item = $(this).find(".selectItemBox:checked").first();
			
			if(item.length > 0){
				selected.push($(this))
			}
			
			
			
        });
		return selected
}

$(function() {
	
	
    $("#reportSelected").button({
        icons: {
            primary: "ui-icon-locked"
        }

    }).click(function(evt) {

		selected = findSelected()
		text = ""
		if(selected.length >0 ){
			for(i=0; i < selected.length;i++){
				descDiv = selected[i].find(".tdDesc").first()
				text += descDiv.html()+"<br>"
			}
			$("#modalDial").html(text)
			$("#modalDial").dialog({
                    maxWidth:600,
                    maxHeight: 500,
                    width: 600,
                    height: 500,
                    modal: true});	
		}
		
		

    });

});