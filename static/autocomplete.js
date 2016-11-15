'use strict';

jQuery(function ($, undefined) {
	var $select = $('<select>', {
		'data-placeholder': "Select a movie"
	}).append('<option>');
	
	$.get(
		'/data',
		function (data, status) {
			if (status !== 'success') {
				console.log("AJAX ERROR");
				return;
			}
			
			$.each(data, function (i, d) {
				var id = d.Id;
				var title = d.Movie.Title;
				
				var $option = $('<option>', {
					'data-id': id
				}).text(title);
				
				$select.append($option);
			});
			
			$('#loading').replaceWith($select);
			
			$select.chosen({
				no_results_text: "No movies match"
			});
		},
		'json'
	);
	
	$select.change(function (e) {
		var $option = $(this).find(':selected');
		var id = $option.attr('data-id');
		window.location = '/movie/'+id;
	})
});
