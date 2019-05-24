$(function () {
    const reqEditor = CodeMirror.fromTextArea(document.getElementById("reqEditor"), {
        mode: 'application/json', lineNumbers: true
    });
    const rspEditor = CodeMirror.fromTextArea(document.getElementById("rspEditor"), {
        mode: 'application/json', lineNumbers: true
    });
    reqEditor.setValue('{}');
    rspEditor.setValue('{}');

    let lastRow = null;

    let routerPath = "";
    let configID = null;
    let MODE = 0;


    const toggleDiv = function () {
        $('#configDiv').toggle(MODE === 2 || MODE === 3);
        if (MODE === 2) {
            $('#btnSend').text("Save Config")
        } else {
            $('#btnSend').text("Test Notify")
        }

        $('#btnDelete').toggle(MODE === 2)
    };

    const highlightCurrentRow = function (td) {
        if (lastRow) {
            lastRow.removeClass('success')
        }
        lastRow = td.parents('tr');
        lastRow.addClass('success');

        routerPath = td.parents('tr').find('.path').text();
    };

    function ajaxError(jqXHR, textStatus, errorThrown) {
        alert(jqXHR.responseText + "\nStatus: " + textStatus + "\nError: " + errorThrown)
    }

    $('.testLink').click(function (event) {
        event.preventDefault();
        highlightCurrentRow($(this));
        configID = null;
        MODE = 1;

        $.ajax({
            type: 'GET',
            url: '/raw' + routerPath,
            success: function (content) {
                reqEditor.setValue(JSON.stringify(content, null, 4));
                rspEditor.setValue('{}');
                toggleDiv()
            },
            error: ajaxError
        })
    });

    $('.editConfig').click(function (event) {
        event.preventDefault();
        highlightCurrentRow($(this));
        configID = $(this).text();

        let editUrl = "";
        if ($(this).hasClass("New")) {
            editUrl = '/config/' + configID + routerPath;
            $('#configIDInput').val("")
        } else {
            editUrl = '/config/' + configID;
            $('#configIDInput').val(configID)
        }

        MODE = 2;

        $.ajax({
            type: 'GET',
            url: editUrl,
            success: function (content) {
                reqEditor.setValue(JSON.stringify(content, null, 4));
                rspEditor.setValue('{}');
                toggleDiv()
            },
            error: ajaxError
        })
    });

    $('.configNotify').click(function (event) {
        event.preventDefault();
        highlightCurrentRow($(this));
        configID = $(this).text();
        $('#configIDInput').val(configID);
        MODE = 3;

        $.ajax({
            type: 'GET',
            url: '/notify/' + configID,
            success: function (content) {
                reqEditor.setValue(JSON.stringify(content, null, 4));
                rspEditor.setValue('{}');
                toggleDiv()
            },
            error: ajaxError
        })
    });

    $('#btnSend').click(function () {
        if (MODE === 1) {
            $.ajax({
                type: 'POST',
                url: '/raw' + routerPath,
                processData: false,
                data: reqEditor.getValue(),
                success: function (content) {
                    rspEditor.setValue(JSON.stringify(content, null, 4))
                },
                error: ajaxError
            })
        } else if (MODE === 2) {
            $.ajax({
                type: 'POST',
                url: '/config/' + $('#configIDInput').val(),
                processData: false,
                data: reqEditor.getValue(),
                success: function (content) {
                    rspEditor.setValue(JSON.stringify(content, null, 4))
                },
                error: ajaxError
            })
        } else if (MODE === 3) {
            $.ajax({
                type: 'POST',
                url: '/notify/' + $('#configIDInput').val(),
                processData: false,
                data: reqEditor.getValue(),
                success: function (content) {
                    rspEditor.setValue(JSON.stringify(content, null, 4))
                },
                error: ajaxError
            })
        }
    });

    $('#btnDelete').click(function () {
        $.ajax({
            type: 'DELETE',
            url: '/config/' + $('#configIDInput').val(),
            success: function (content) {
                rspEditor.setValue(JSON.stringify(content, null, 4))
                document.location.reload()
            },
            error: ajaxError
        })
    });
})