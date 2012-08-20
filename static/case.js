// JS for case pages

var CaseSetup = function(caseId) {
    $(function() {
        $(".notes input, .notes textarea").change(function() {
            var values = {
                "of-interest"  : $("#of-interest-chbx").attr("checked") == "checked",
                "notes"        : $("#notes-txt").val(),
                "clothing"     : $("#clothing-chbx").attr("checked") == "checked",
                "raw-textiles" : $("#raw-textiles-chbx").attr("checked") == "checked",
                "other"        : $("#other-chbx").attr("checked") == "checked"
            };
            // send a request to update the record
            $.post("/case/" + caseId, values);          
        });
    });
};
