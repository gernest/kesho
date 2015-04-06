/**
 * Created by gernest on 3/25/15.
 */

$(document).ready(function(){
    $('.js-flash-close')
        .on('click', function(){
            $(this).closest('.flash').remove()
        })
});