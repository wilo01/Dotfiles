








ca_printer_dimension

    select skp.* , b.forename, b.surname
       from
          safe_kiosk_print skp, ca_booked_person b
       where skp.collection_point_code in ('129', '130')
       and skp.booking_reference = b.booking_reference
       and skp.printed_date_time is null
       and skp.requested_date_time > sysdate - 1
       and b.due_date_time > sysdate -1
       order by decode( b.booking_source_ind, 'Q', b.booking_reference) , b.surname, b.forename, skp.job_number


{"width":"560", "height":"480"} 1
{"width":"500", "height":"410"} 1.13


update ca_printer_dimension set bluetooth_pixel_dimensions = '{"width":"500", "height":"410"}', bluetooth_scale_factor = '1.13' where printer_code = 'star';