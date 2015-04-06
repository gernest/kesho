
$(document).ready(function(){
    var editor=new Editor();
    editor.render();
    html5tooltips([
        {
            contentText:"Inakozesha maneno",
            targetSelector:".icon-bold",
            stickTo:'top'
        },
        {
            contentText:"Staili ya mlazo",
            targetSelector: ".icon-italic",
            stickTo:'top'
        },
        {
            contentText:"Onyesha msisitizo",
            targetSelector: ".icon-quote",
            stickTo:"top"
        },
        {
            contentText:"Orodha isiyokuwa na namba",
            targetSelector:".icon-unordered-list",
            stickTo:"top"
        },
        {
            contentText:"Orodha yenye namba",
            targetSelector:".icon-ordered-list",
            stickTo:"top"
        },
        {
            contentText:"Linki",
            targetSelector:".icon-link",
            stickTo:"top"
        },
        {
            contentText:"Cheza",
            targetSelector:".icon-play",
            stickTo:"top"
        },
        {
            contentText:"muziki",
            targetSelector:".icon-music",
            stickTo:"top"
        },
        {
            contentText:"weka picha",
            targetSelector:".icon-image",
            stickTo:"top"
        },
        {
            contentText:"Mkataba",
            targetSelector:".icon-contract",
            stickTo:"top"
        },
        {
            contentText:"kuza eneo ka kuandikia",
            targetSelector:".icon-fullscreen",
            stickTo:"top"
        },
        {
            contentText:"Uliza Swali",
            targetSelector:".icon-question",
            stickTo:"top"
        },
        {
            contentText:"Ufafanuzi",
            targetSelector:".icon-info",
            stickTo:"top"
        },
        {
            contentText:"Batili mabadiliko uliyo fanya, rudi nyuma",
            targetSelector:".icon-undo",
            stickTo:"top"
        },
        {
            contentText:"Batili mabadiliko uliyofanya, nenda mbele",
            targetSelector:".icon-redo",
            stickTo:"top"
        },
        {
            contentText:"Code",
            targetSelector:".icon-code",
            stickTo:"top"
        },
        {
            contentText:"Angalia makala yako jinsi itakavyoonekana utakapotuma",
            targetSelector:".icon-preview",
            stickTo:"top"
        }
    ])
})