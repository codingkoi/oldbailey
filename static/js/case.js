// JS for case pages

var CaseSetup = function(caseId) {
    $(function() {
        $(".notes input, .notes textarea").change(function() {
            var values = {
                "OfInterest"    : $("#of-interest-chbx").attr("checked") == "checked",
                "NotOfInterest" : $("#not-of-interest-chbx").attr("checked") == "checked",
                "Notes"         : $("#notes-txt").val(),
                "Clothing"      : $("#clothing-chbx").attr("checked") == "checked",
                "ClothingCount" : parseInt($("#clothing-cnt").val()) || 0,
                "RawTextiles"   : $("#raw-textiles-chbx").attr("checked") == "checked",
                "RawTextilesCount" : parseInt($("#raw-textiles-cnt").val()) || 0,
                "OtherTextiles" : $("#other-textiles-chbx").attr("checked") == "checked",
                "HouseholdLinen" : $("#household-linen-chbx").attr("checked") == "checked",
                "HouseholdLinenCount" : parseInt($("#household-linen-cnt").val()) || 0,
                "Accessories" : $("#accessories-chbx").attr("checked") == "checked",
                "AccessoriesCount" : parseInt($("#accessories-cnt").val()) || 0,
                "Other"         : $("#other-chbx").attr("checked") == "checked",
                "OtherCount" : parseInt($("#other-cnt").val()) || 0,
                "OtherNotSpecified" : $("#other-not-specified-chbx").attr("checked") == "checked"
            };
            // send a request to update the record
            $.post("/case/" + caseId, {"json": JSON.stringify(values)});
        });
    });
};
