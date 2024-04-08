import tippy from '../lib/tippy/dist/tippy.esm.js'


$(document).ready(function () {
    let url = window.location.href.replace(new RegExp("/view" + "$"), "");
    let calendar = new FullCalendar.Calendar(document.getElementById('calendar'), {
        themeSystem: 'bootstrap',
        headerToolbar: {
            left: 'prev,next today',
            center: 'title',
            right: 'dayGridMonth,listWeek,listYear'
        },

        // customize the button names,
        // otherwise they'd all just say "list"
        views: {
            listWeek: {buttonText: 'list week'},
            list: {buttonText: 'list year'}
        },

        eventDidMount: function (info) {
            let title = info.event.extendedProps.description;
            title = title.replaceAll('\n', ' <br/> ');
            title = title.replace(/(https?:\/\/)\S+/g, function (matched) {
                return `<a href="${matched}" target="_blank" rel="noopener noreferrer">${matched}</a>`;
            });
            info.event.setProp('url', '');
            tippy(info.el, {
                placement: 'top',
                content: title,
                arrow: true,
                // trigger: 'click focus',
                trigger: 'mouseenter',
                allowHTML: true,
                interactive: true,
                theme: 'translucent',
                maxWidth: 600,
            });
        },
        eventClick: function (info) {
            info.jsEvent.preventDefault();
        },
        eventDisplay: 'block',
        eventTimeFormat: {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            meridiem: false,
            hour12: false,
        },
        displayEventEnd: false,

        initialView: 'dayGridMonth',
        navLinks: true, // can click day/week names to navigate views
        editable: false,
        dayMaxEvents: true, // allow "more" link when too many events
        nowIndicator: true,
        defaultTimedEventDuration: '00:00:01',
        events: {
            url: url,
            format: 'ics'
        },
        eventSourceSuccess: function (_, response) {
            let h = response.headers.get("content-disposition");
            h = h.split('"')[1]
            $('#navbar-center-content').html(`<h2>Calendar: ${h}</h2>`);
        }
    });

    calendar.render();
});
