class OldBaileyModel
    constructor: (data) ->
        @results = ko.mapping.fromJS(data)
        @curCase = ko.observable @results.Records()[0]
    newSearch: ->
        $.get "/search?text=#{@results.SearchText()}", @update, "json"
    update: (data) =>
        ko.mapping.fromJS(data, @results)
    backToResults: ->
        @curCase null
    selectCase: (caseObj) =>
        @curCase caseObj 

# initial load
callback = (data) ->
    window.oldBaileyModel = new OldBaileyModel(data)
    ko.applyBindings window.oldBaileyModel
$.get "/search", callback, "json"