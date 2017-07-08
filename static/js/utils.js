function handleTabSwitch(responseContainer, responseArea) {
    var target = $(this).attr('href');

    // switch tabs
    $('#tab_content .tab:visible').addClass('hide')
        .removeClass('is_active');

    $(target).toggleClass('is_active').toggleClass('hide');

    $('#tab_links a').removeClass('is-active');
    $(this).addClass('is-active');

    // clean up old errors or results
    responseContainer.addClass('mdl-color--yellow-100')
        .removeClass('mdl-color--red-200')
        .removeClass('mdl-color--green-200');

    responseArea.html('');

    if ($(this).attr('data-on-load')) {
        var func = window[$(this).attr('data-on-load')];
        if (typeof func === 'function') {
            func();
        }
    }
}