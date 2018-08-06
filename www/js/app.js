var $$ = Dom7;

var DaikinWeb = {
    controlCard: null,
    units: {},
    refreshTimeout: null,
    unitTemplate: undefined,
    refresh: (done) => {
	// update already running
	if (DaikinWeb.updateRunning) {
	    if (done) done();
	    return;
	}

	DaikinWeb.updateRunning = true;

	if (DaikinWeb.refreshTimeout) {
	    clearTimeout(DaikinWeb.refreshTimeout);
	}

	Utils.APIRequest('/units', {
	    success: (result, status, xhr) => {
		let controlinfo = {
		    pow: null,
		    stemp: null,
		    shum: null,
		    f_dir: null,
		    f_rate: null,
		    mode: null,
		    otemp: null,
		};
		result.data.forEach(unit => {
		    if (!unit.name) return;

		    DaikinWeb.updateUnit(unit);

		    // collect control info from all units
		    for (let key in controlinfo) {
			if (!controlinfo.hasOwnProperty(key)) continue;

			let val = controlinfo[key];

			// otemp should not be different,
			// but it is sometimes
			if (val !== null && val != unit[key] && key !== 'otemp') {
			    val = "mixed";
			} else if (val === null) {
			    val = unit[key];
			}

			controlinfo[key] = val;
		    }
		});

		DaikinWeb.updateControl(controlinfo);

		DaikinWeb.updateRunning = false;
		DaikinWeb.refreshTimeout = setTimeout(() => {
		    DaikinWeb.refresh();
		}, 5000);

		if (done) done();
	    },
	    error: (xhr, error) => {
		app.toast.create({
		    title: "Error",
		    text: error,
		}).open();

		DaikinWeb.refreshTimeout = setTimeout(() => {
		    DaikinWeb.refresh();
		}, 5000);

		if (done) done();
	    },
	});
    },
    updateUnit: (data) => {
	let context = {
	    mode_icon: Utils.mode_icon[data.mode],
	    title: Utils.render_unit_title(data),
	    togglestate: (data.pow == 1) ? 'checked' : '',
	    name: data.name,
	    stemp: data.stemp,
	    dir_icon: Utils.f_dir_icon[data.f_dir],
	    rate_icon: Utils.f_rate_icon[data.f_rate],
	};

	let el =  Dom7(DaikinWeb.unitTemplate(context));
	el.data('control', data);

	// Setup Toggle
	app.toggle.create({
	    el: el.find('label[name=pow]'),
	    on: {
		change: (toggle) => {
		    DaikinWeb.setUnitPower(toggle, el);
		}
	    }
	});

	let oldItem = DaikinWeb.units[data.name];
	if (oldItem) {
	    el.insertAfter(oldItem);
	    oldItem.remove();
	} else {
	    $$('#mainpage').append(el);
	}
	DaikinWeb.units[data.name] = el;
    },
    updateControl: (data, pageEl) => {
	// setup data
	data.togglestate = (data.pow == 1) ? 'checked' : '';

	let oldEl = DaikinWeb.controlCard;
	let el = Dom7(DaikinWeb.controlTemplate(data));
	el.data('control', data);
	// setup handlers

	app.toggle.create({
	    el: el.find('.toggle'),
	    on: {
		change: (toggle) => {
		    DaikinWeb.setUnitPower(toggle, el);
		}
	    }
	});

	app.stepper.create({
	    el: el.find('.stepper'),
	    formatValue: Utils.render_temperature,
	    on: {
		change: (stepper, value) => {
		    DaikinWeb.setTargetTemperature(value, el);
		},
	    },
	});

	['mode', 'f_dir', 'f_rate'].forEach((type) => {
	    let modeButtons = el.find('.segmented[name=' + type + '] button');
	    let activeMode = modeButtons.filter((index, el) => {
		return el.name === data[type];
	    });
	    activeMode.addClass('button-active');
	    modeButtons.on('click', function(e) {
		var me = Dom7(this);
		let value = me.attr('name');
		DaikinWeb.setControlParameter(type, value, el);
		modeButtons.removeClass('button-active');
		me.addClass('button-active');
	    });
	});

	if (pageEl) {
	    pageEl.empty();
	    pageEl.append(el);
	    return;
	} else if (oldEl) {
	    el.insertBefore(oldEl);
	    oldEl.remove();
	} else {
	    el.insertAfter('.ptr-preloader');
	}
	DaikinWeb.controlCard = el;
    },
    setUnitPower: (toggle, unit) => {
	let data = unit.data('control');
	if ( (data.pow == 1) === toggle.checked ) {
	    // the toggle was changed by us, not the user
	    return;
	}

	data.pow = toggle.checked ? '1' : '0';

	Utils.APIRequest(Utils.generate_control_url(data), {
	    method: 'PUT',
	    params: {
		pow: data.pow
	    }
	});
    },
    setTargetTemperature: (value, unit) => {
	let data = unit.data('control');
	data.stemp = value;
	Utils.APIRequest(Utils.generate_control_url(data), {
	    method: 'PUT',
	    params: {
		stemp: data.stemp
	    }
	});
    },
    setControlParameter: (type,value,unit) => {
	let data = unit.data('control');
	data[type] = value;

	let params = {};
	params[type] = value;

	Utils.APIRequest(Utils.generate_control_url(data), {
	    method: 'PUT',
	    params: params,
	});
    }
};

var Utils = {
    mode_icon: {
	0: 'brightness_auto',
	1: 'brightness_auto',
	7: 'brightness_auto',
	2: 'opacity',
	3: 'ac_unit',
	4: 'wb_sunny',
	6: 'toys',
    },
    f_dir_icon: {
	0: 'block',
	1: 'swap_vert',
	2: 'swap_horiz',
	3: '3d_rotation',
    },
    f_rate_icon: {
	A: 'brightness_auto',
	B: 'volume_off',
	3: 'looks_one',
	4: 'looks_two',
	5: 'looks_3',
	6: 'looks_4',
	7: 'looks_5',
    },
    render_unit_title: (data) => {
	return data.name + ' - ' + data.htemp + '°';
    },
    render_temperature: (val) => val.toFixed(1),
    generate_form_data: (data, givenFields) => {
	let defaultFields = ['pow', 'mode', 'stemp', 'shum', 'f_dir', 'f_rate'];
	let fields = givenFields || defaultFields;
	let formData = {};
	fields.forEach(field => {
	    formData[field] = data[field];
	});

	return formData;
    },
    generate_control_url: (data) => {
	let url = "/units";
	if (data.name) {
	    url += '/' + data.name + '/control';
	}

	return url;
    },
    APIRequest: (url, options) => {
	// options:
	// 

	let error_cb = options.error || function(xhr, status) {
	    let toast = app.toast.create({
		title: "Error",
		text: status,
	    });
	    toast.open();
	};

	app.request({
	    url: url,
	    method: options.method || "GET",
	    data: options.params,
	    dataType: "json",
	    error: error_cb,
	    success: options.success,
	    complete: options.complete,
	});

    },
};

var app = new Framework7({
    root: '#app',
    name: 'Daikin Web',
    id: 'com.csapak.daikinweb',
    theme: 'auto',
    panel: {
	swipe: 'left',
    },
    routes: [
	{
	    name: 'about',
	    path: '/about/',
	    url: '/about.html',
	    on: {
		pageBeforeIn: (event, page) => {
		    console.log(event,page);
		    app.panel.left.close();
		},
	    },
	},
	{
	    name: 'unit',
	    path: '/units/:name/',
	    templateUrl: '/control.html',
	    on: {
		pageInit: (event, page) => {
		    let unit = DaikinWeb.units[page.name];
		    let updatePage = (data) => {
			data.name = data.name || page.name;
			DaikinWeb.updateControl(data, page.$el.find('.page-content'))
			app.preloader.hide();
		    };
		    if (!unit) {
			app.preloader.show();
			Utils.APIRequest('/units/' + page.name + '/control', {
			    success: updatePage
			});
		    } else {
			updatePage(unit.data('control'));
		    }
		},
	    },
	},
    ],
    toast: {
	closeButton: true,
    },
});

var mainView = app.views.create('.view-main', {
    url: '/',
    pushState: true,
});

var ptr = app.ptr.get('.ptr-content');;

ptr.$el.on('ptr:refresh', e => {
    DaikinWeb.refresh(app.ptr.done);
});

// compile templates
DaikinWeb.unitTemplate = Template7.compile($$('#unit-template').html());
DaikinWeb.controlTemplate = Template7.compile($$('#control-template').html());

// first refresh
DaikinWeb.refreshTimeout = setTimeout(() => { ptr.refresh(); }, 1);
