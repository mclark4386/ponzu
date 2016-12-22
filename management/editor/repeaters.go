package editor

import (
	"bytes"
	"fmt"
	"strings"
)

// InputRepeater returns the []byte of an <input> HTML element with a label.
// It also includes repeat controllers (+ / -) so the element can be
// dynamically multiplied or reduced.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
// 	type Person struct {
// 		Name string `json:"name"`
// 	}
//
// 	func (p *Person) MarshalEditor() ([]byte, error) {
// 		view, err := editor.Form(p,
// 			editor.Field{
// 				View: editor.InputRepeater("Name", p, map[string]string{
// 					"label":       "Name",
// 					"type":        "text",
// 					"placeholder": "Enter the Name here",
// 				}),
// 			}
// 		)
// 	}
func InputRepeater(fieldName string, p interface{}, attrs map[string]string) []byte {
	// find the field values in p to determine pre-filled inputs
	fieldVals := valueFromStructField(fieldName, p)
	vals := strings.Split(fieldVals, "__ponzu")

	scope := tagNameFromStructField(fieldName, p)
	html := bytes.Buffer{}

	html.WriteString(`<span class="__ponzu-repeat ` + scope + `">`)
	for i, val := range vals {
		el := &element{
			TagName: "input",
			Attrs:   attrs,
			Name:    tagNameFromStructFieldMulti(fieldName, i, p),
			data:    val,
			viewBuf: &bytes.Buffer{},
		}

		// only add the label to the first input in repeated list
		if i == 0 {
			el.label = attrs["label"]
		}

		html.Write(domElementSelfClose(el))
	}
	html.WriteString(`</span>`)

	return append(html.Bytes(), RepeatController(fieldName, p, "input", ".input-field")...)
}

// SelectRepeater returns the []byte of a <select> HTML element plus internal <options> with a label.
// It also includes repeat controllers (+ / -) so the element can be
// dynamically multiplied or reduced.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func SelectRepeater(fieldName string, p interface{}, attrs, options map[string]string) []byte {
	// options are the value attr and the display value, i.e.
	// <option value="{map key}">{map value}</option>
	scope := tagNameFromStructField(fieldName, p)
	html := bytes.Buffer{}
	html.WriteString(`<span class="__ponzu-repeat ` + scope + `">`)

	// find the field values in p to determine if an option is pre-selected
	fieldVals := valueFromStructField(fieldName, p)
	vals := strings.Split(fieldVals, "__ponzu")

	attrs["class"] = "browser-default"

	// loop through vals and create selects and options for each, adding to html
	if len(vals) > 0 {
		for i, val := range vals {
			sel := &element{
				TagName: "select",
				Attrs:   attrs,
				Name:    tagNameFromStructFieldMulti(fieldName, i, p),
				viewBuf: &bytes.Buffer{},
			}

			// only add the label to the first select in repeated list
			if i == 0 {
				sel.label = attrs["label"]
			}

			// create options for select element
			var opts []*element

			// provide a call to action for the select element
			cta := &element{
				TagName: "option",
				Attrs:   map[string]string{"disabled": "true", "selected": "true"},
				data:    "Select an option...",
				viewBuf: &bytes.Buffer{},
			}

			// provide a selection reset (will store empty string in db)
			reset := &element{
				TagName: "option",
				Attrs:   map[string]string{"value": ""},
				data:    "None",
				viewBuf: &bytes.Buffer{},
			}

			opts = append(opts, cta, reset)

			for k, v := range options {
				optAttrs := map[string]string{"value": k}
				if k == val {
					optAttrs["selected"] = "true"
				}
				opt := &element{
					TagName: "option",
					Attrs:   optAttrs,
					data:    v,
					viewBuf: &bytes.Buffer{},
				}

				opts = append(opts, opt)
			}

			html.Write(domElementWithChildrenSelect(sel, opts))
		}
	}

	html.WriteString(`</span>`)
	return append(html.Bytes(), RepeatController(fieldName, p, "select", ".input-field")...)
}

// FileRepeater returns the []byte of a <input type="file"> HTML element with a label.
// It also includes repeat controllers (+ / -) so the element can be
// dynamically multiplied or reduced.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func FileRepeater(fieldName string, p interface{}, attrs map[string]string) []byte {
	// find the field values in p to determine if an option is pre-selected
	fieldVals := valueFromStructField(fieldName, p)
	vals := strings.Split(fieldVals, "__ponzu")

	addLabelFirst := func(i int, label string) string {
		if i == 0 {
			return `<label class="active">` + label + `</label>`
		}

		return ""
	}

	tmpl :=
		`<div class="file-input %[5]s %[4]s input-field col s12">
			%[2]s
			<div class="file-field input-field">
				<div class="btn">
					<span>Upload</span>
					<input class="upload %[4]s" type="file" />
				</div>
				<div class="file-path-wrapper">
					<input class="file-path validate" placeholder="Add %[5]s" type="text" />
				</div>
			</div>
			<div class="preview"><div class="img-clip"></div></div>			
			<input class="store %[4]s" type="hidden" name="%[1]s" value="%[3]s" />
		</div>`
		// 1=nameidx, 2=addLabelFirst, 3=val, 4=className, 5=fieldName
	script :=
		`<script>
			$(function() {
				var $file = $('.file-input.%[2]s'),
					upload = $file.find('input.upload'),
					store = $file.find('input.store'),
					preview = $file.find('.preview'),
					clip = preview.find('.img-clip'),
					reset = document.createElement('div'),
					img = document.createElement('img'),
					uploadSrc = store.val();
					preview.hide();
				
				// when %[2]s input changes (file is selected), remove
				// the 'name' and 'value' attrs from the hidden store input.
				// add the 'name' attr to %[2]s input
				upload.on('change', function(e) {
					resetImage();
				});

				if (uploadSrc.length > 0) {
					$(img).attr('src', store.val());
					clip.append(img);
					preview.show();

					$(reset).addClass('reset %[2]s btn waves-effect waves-light grey');
					$(reset).html('<i class="material-icons tiny">clear<i>');
					$(reset).on('click', function(e) {
						e.preventDefault();
						var preview = $(this).parent().closest('.preview');
						preview.animate({"opacity": 0.1}, 200, function() {
							preview.slideUp(250, function() {
								resetImage();
							});
						})
						
					});
					clip.append(reset);
				}

				function resetImage() {
					store.val('');
					store.attr('name', '');
					upload.attr('name', '%[1]s');
					clip.empty();
				}
			});	
		</script>`
		// 1=nameidx, 2=className

	name := tagNameFromStructField(fieldName, p)

	html := bytes.Buffer{}
	html.WriteString(`<span class="__ponzu-repeat ` + name + `">`)
	for i, val := range vals {
		className := fmt.Sprintf("%s-%d", name, i)
		nameidx := tagNameFromStructFieldMulti(fieldName, i, p)
		html.WriteString(fmt.Sprintf(tmpl, nameidx, addLabelFirst(i, attrs["label"]), val, className, fieldName))
		html.WriteString(fmt.Sprintf(script, nameidx, className))
	}
	html.WriteString(`</span>`)

	return append(html.Bytes(), RepeatController(fieldName, p, "input.upload", "div.file-input."+fieldName)...)
}

// RepeatController generates the javascript to control any repeatable form
// element in an editor based on its type, field name and HTML tag name
func RepeatController(fieldName string, p interface{}, inputSelector, cloneSelector string) []byte {
	scope := tagNameFromStructField(fieldName, p)
	script := `
    <script>
        $(function() {
            // define the scope of the repeater
            var scope = $('.__ponzu-repeat.` + scope + `');

            var getChildren = function() {
                return scope.find('` + cloneSelector + `')
            }

            var resetFieldNames = function() {
                // loop through children, set its name to the fieldName.i where
                // i is the current index number of children array
                var children = getChildren();

                for (var i = 0; i < children.length; i++) {
					var preset = false;					
                    var $el = children.eq(i);
					var name = '` + scope + `.'+String(i);

                    $el.find('` + inputSelector + `').attr('name', name);

					// ensure no other input-like elements besides ` + inputSelector + `
					// get the new name by setting it to an empty string
					$el.find('input, select, textarea').each(function(i, elem) {
						var $elem = $(elem);
						
						// if the elem is not ` + inputSelector + ` and has no value 
						// set the name to an empty string
						if (!$elem.is('` + inputSelector + `')) {
							if ($elem.val() === '') {
								$elem.attr('name', '');
							} else {
								preset = true;
							}						
						}
					});      

					// if there is a preset value, remove the name attr from the
					// ` + inputSelector + ` element so it doesn't overwrite db
					if (preset) {
						$el.find('` + inputSelector + `').attr('name', '');														
					}          

                    // reset controllers
                    $el.find('.controls').remove();
                }

                applyRepeatControllers();
            }

            var addRepeater = function(e) {
                e.preventDefault();
                
                var add = e.target;

                // find and clone the repeatable input-like element
                var source = $(add).parent().closest('` + cloneSelector + `');
                var clone = source.clone();

                // if clone has label, remove it
                clone.find('label').remove();
                
                // remove the pre-filled value from clone
                clone.find('` + inputSelector + `').val('');
				clone.find('input').val('');

                // remove controls from clone if already present
                clone.find('.controls').remove();

				// remove input preview on clone if copied from source
				clone.find('.preview').remove();

                // add clone to scope and reset field name attributes
                scope.append(clone);

                resetFieldNames();
            }

            var delRepeater = function(e) {
                e.preventDefault();

                // do nothing if the input is the only child
                var children = getChildren();
                if (children.length === 1) {
                    return;
                }

                var del = e.target;
                
                // pass label onto next input-like element if del 0 index
                var wrapper = $(del).parent().closest('` + cloneSelector + `');
                if (wrapper.find('` + inputSelector + `').attr('name') === '` + scope + `.0') {
                    wrapper.next().append(wrapper.find('label'))
                }
                
                wrapper.remove();

                resetFieldNames();
            }

            var createControls = function() {
                // create + / - controls for each input-like child element of scope
                var add = $('<button>+</button>');
                add.addClass('repeater-add');
                add.addClass('btn-flat waves-effect waves-green');

                var del = $('<button>-</button>');
                del.addClass('repeater-del');
                del.addClass('btn-flat waves-effect waves-red');                

                var controls = $('<span></span>');
                controls.addClass('controls');
                controls.addClass('right');

                // bind listeners to child's controls
                add.on('click', addRepeater);
                del.on('click', delRepeater);

                controls.append(add);
                controls.append(del);

                return controls;
            }

            var applyRepeatControllers = function() {
                // add controls to each child
                var children = getChildren()
                for (var i = 0; i < children.length; i++) {
                    var el = children[i];
                    
                    $(el).find('` + inputSelector + `').parent().find('.controls').remove();
                    
                    var controls = createControls();                                        
                    $(el).append(controls);
                }
            }

            applyRepeatControllers();
        });

    </script>
    `

	return []byte(script)
}